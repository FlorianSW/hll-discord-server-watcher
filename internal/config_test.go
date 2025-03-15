package internal_test

import (
	"github.com/floriansw/hll-discord-server-watcher/internal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log/slog"
	"os"
)

var _ = Describe("Config", func() {
	Describe("Persistence", func() {
		It("persists config change of server", func() {
			l := slog.New(slog.NewTextHandler(os.Stdout, nil))
			f, err := os.CreateTemp(os.TempDir(), "config")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())
			Expect(os.WriteFile(f.Name(), []byte("{}"), 0655)).ToNot(HaveOccurred())
			c, err := internal.NewConfig(f.Name(), l)
			Expect(err).ToNot(HaveOccurred())

			Expect(c.Save()).ToNot(HaveOccurred())
			c, err = internal.NewConfig(f.Name(), l)
			Expect(err).ToNot(HaveOccurred())

			Expect(c.Save()).ToNot(HaveOccurred())

			c, err = internal.NewConfig(f.Name(), l)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
