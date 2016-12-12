package xenstatsd

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
	PollInterval  int          `json:"poll_interval"`  // Time in seconds between each poll of metrics
	RetryInterval int          `json:"retry_interval"` // Time in seconds between retry for connection
	Hosts         []HostConfig `json:hosts`
	StatsdHost    string       `json:"statsd_host"`
	StatsdPort    int          `json:"statsd_port"`
	StatsdPrefix  string       `json:"statsd_prefix"`
}

type HostConfig struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type StatsdMetric struct {
	Name  string
	Value string
}
