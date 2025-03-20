package watcher

import (
	"github.com/floriansw/go-tcadmin"
	"github.com/floriansw/hll-discord-server-watcher/internal"
)

type ServerQuery interface {
	ServerInfo(serviceId string) (*tcadmin.ServerInfo, error)
}

type Server struct {
	Query  ServerQuery
	Config internal.Server
}
