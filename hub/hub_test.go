package hub

import "net"

var _ net.Conn = &ConnectionHub{}
