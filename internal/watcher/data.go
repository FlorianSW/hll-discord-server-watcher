package watcher

import (
	"github.com/floriansw/hll-discord-server-watcher/internal"
	"github.com/floriansw/hll-discord-server-watcher/tcadmin"
)

type ServerQuery interface {
	ServerInfo(serviceId string) (*tcadmin.ServerInfo, error)
}

type Server struct {
	Query  ServerQuery
	Config internal.Server
}
