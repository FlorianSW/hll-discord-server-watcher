package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/floriansw/hll-discord-server-watcher/discord"
	"github.com/floriansw/hll-discord-server-watcher/internal"
	"github.com/floriansw/hll-discord-server-watcher/internal/watcher"
	"github.com/floriansw/hll-discord-server-watcher/tcadmin"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	level := slog.LevelInfo
	if _, ok := os.LookupEnv("DEBUG"); ok {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	c, err := internal.NewConfig("./config.json", logger)
	if err != nil {
		logger.Error("config", err)
		return
	}

	var s *discordgo.Session
	if c.Discord != nil {
		s, err = discordgo.New("Bot " + c.Discord.Token)
		if err != nil {
			logger.Error("discord", err)
			return
		}
	}
	if err = os.MkdirAll("./matches/", 0644); err != nil {
		logger.Error("create-matches", err)
		return
	}
	h := discord.New(logger, c, s)
	if s != nil {
		s.AddHandlerOnce(func(s *discordgo.Session, e *discordgo.Ready) {
			if err := h.Listen(); err != nil {
				logger.Error("discord-listen", err)
				panic(err)
			}
			logger.Info("ready")
		})
		err = s.Open()
		if err != nil {
			logger.Error("open-session", err)
			return
		}
		defer s.Close()
	}
	defer h.Close()

	var servers []watcher.Server
	for _, server := range c.Servers {
		jar, err := cookiejar.New(nil)
		if err != nil {
			panic(err)
		}
		hc := http.Client{
			Jar: jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		servers = append(servers, watcher.Server{
			Query: tcadmin.NewClient(hc, "qp.qonzer.com", tcadmin.Credentials{
				Username: server.Credentials.Username,
				Password: server.Credentials.Password,
			}),
			Config: server,
		})
	}
	interval := 10 * time.Minute
	if c.PollIntervalSeconds != nil {
		interval = time.Duration(*c.PollIntervalSeconds) * time.Second
	}
	watcher.NewWatcher(logger, s, c, servers, interval).Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("graceful-shutdown")
	if err := c.Save(); err != nil {
		logger.Error("save-config", "error", err)
	}
}
