package xengraphite

import (
	"github.com/nilshell/xmlrpc"
)

type XenApiClient struct {
	Session  interface{}
	Url      string
	Username string
	Password string
	Rpc      *xmlrpc.Client
}

type Config struct {
	PollInterval int          `json:"poll_interval"`
	Hosts        []HostConfig `json:hosts`
	GraphiteHost string       `json:"graphite_host"`
	GraphitePort int          `json:"graphite_port"`
}

type HostConfig struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}
