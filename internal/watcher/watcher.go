package watcher

import (
	"github.com/bwmarrin/discordgo"
	"github.com/floriansw/hll-discord-server-watcher/internal"
	"log/slog"
	"time"
)

type watcher struct {
	logger  *slog.Logger
	servers []Server
	s       *discordgo.Session
	c       *internal.Config

	ticker *time.Ticker
}

func NewWatcher(l *slog.Logger, s *discordgo.Session, c *internal.Config, servers []Server, d time.Duration) *watcher {
	return &watcher{
		logger:  l,
		servers: servers,
		ticker:  time.NewTicker(d),
		s:       s,
		c:       c,
	}
}

func (w *watcher) Run() {
	go w.watchServers()
}

type serverInfo struct {
	Name           string
	Color          *int
	ServerName     string
	ServerPassword string
}

func (w *watcher) watchServers() {
	for {
		select {
		case <-w.ticker.C:
			var servers []serverInfo
			for _, server := range w.servers {
				si, err := server.Query.ServerInfo(server.Config.ServiceId)
				if err != nil {
					w.logger.Error("server-query", "server", server.Config.Name, "error", err)
					servers = append(servers, serverInfo{Name: server.Config.Name})
					continue
				}
				servers = append(servers, serverInfo{
					Name:           server.Config.Name,
					Color:          server.Config.Color,
					ServerName:     si.Name,
					ServerPassword: si.Password,
				})
			}
			go w.publish(servers)
		}
	}
}

func (w *watcher) publish(s []serverInfo) {
	if w.c.Discord.MessageId == nil {
		w.createMessage(s)
	} else {
		w.updateMessage(s)
	}
}

func (w *watcher) createMessage(s []serverInfo) {
	message, err := w.s.ChannelMessageSendComplex(w.c.Discord.ChannelId, &discordgo.MessageSend{
		Embeds: serverStatus(s),
	})
	if err != nil {
		w.logger.Error("create-message", "error", err)
	} else {
		w.c.Discord.MessageId = &message.ID
	}
}

func (w *watcher) updateMessage(s []serverInfo) {
	embeds := serverStatus(s)
	message, err := w.s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Embeds:  &embeds,
		ID:      *w.c.Discord.MessageId,
		Channel: w.c.Discord.ChannelId,
	})

	if err != nil {
		w.logger.Error("create-message", "error", err)
	} else {
		w.c.Discord.MessageId = &message.ID
	}
}

func serverStatus(s []serverInfo) (embeds []*discordgo.MessageEmbed) {
	for _, info := range s {
		color := internal.ColorDarkGrey
		if info.Color != nil {
			color = *info.Color
		}
		embeds = append(embeds, &discordgo.MessageEmbed{
			Title: info.Name,
			Color: color,
			Fields: []*discordgo.MessageEmbedField{{
				Name:  "Server Name",
				Value: info.ServerName,
			}, {
				Name:  "Password",
				Value: info.ServerPassword,
			}},
		})
	}
	return
}
