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

type Metric struct {
	Timestamp string
	Key       string
	Value     string
}
