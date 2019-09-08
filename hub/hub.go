package hub

import (
	"time"

	"context"
	"io"
	"net"
	"sync"

	hub "github.com/shellimsi/proto/hub"
)

var (
	defTimeout = time.Second * 5
)

type ConnectionHub struct {
	connOnHub    hub.ConnectionClient
	readTimeout  time.Duration
	writeTimeout time.Duration
	index        int64
	sync.Mutex
}

func New(connHub hub.ConnectionClient) (conn *ConnectionHub, err error) {
	return &ConnectionHub{
		connOnHub:    connHub,
		readTimeout:  defTimeout,
		writeTimeout: defTimeout,
		index:        0,
	}, nil
}

func (c *ConnectionHub) Read(b []byte) (n int, err error) {
	c.Lock()
	defer c.Unlock()
	rSize := uint32(len(b))
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(c.readTimeout))
	defer cancel()
	res, err := c.connOnHub.Read(ctx, &hub.ReadRequest{
		Size: rSize,
	}, nil)

	if err != nil {
		panic(err)
	}

	switch res.GetErr() {
	case hub.ConnectionErr_EOF:

		return int(res.GetSize()), io.EOF
	case hub.ConnectionErr_SHORTWRITE:
		return 0, io.ErrShortWrite
	case hub.ConnectionErr_UNEXPECTEDEOF:
		return 0, io.ErrUnexpectedEOF
	case hub.ConnectionErr_CLOSEDPIPE:
		return 0, io.ErrClosedPipe
	}
	n = copy(b, res.GetData())

	return n, nil
}

func (c *ConnectionHub) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (c *ConnectionHub) Close() error {
	return nil
}

func (c *ConnectionHub) LocalAddr() net.Addr {
	var addr net.Addr
	return addr
}

func (c *ConnectionHub) RemoteAddr() net.Addr {
	var addr net.Addr
	return addr
}

func (c *ConnectionHub) SetDeadline(t time.Time) error {
	return nil

}

func (c *ConnectionHub) SetReadDeadline(t time.Time) error {
	return nil

}

func (c *ConnectionHub) SetWriteDeadline(t time.Time) error {
	return nil

}
