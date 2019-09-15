package hub

import (
	"errors"
	"time"

	"context"
	"io"
	"net"
	"sync"

	uuid "github.com/satori/go.uuid"
	hub "github.com/shellimsi/proto/hub"
)

var (
	ErrAgentNotRegistered = errors.New("the agent could not register  Hub Server")
	ErrAlreadyClosed      = errors.New("already closed")
	defTimeout            = time.Second * 5
	defaultOptions        = options{
		againtID: uuid.NewV4().String(),
		timeOut:  time.Second * 10,
	}
)

type options struct {
	againtID string
	timeOut  time.Duration
}

type ConnectionOptions func(*options)

func AgentID(agentID uuid.UUID) ConnectionOptions {
	return func(o *options) {
		o.againtID = agentID.String()
	}
}

func TimeOut(duration time.Duration) ConnectionOptions {
	return func(o *options) {
		o.timeOut = duration
	}
}

type ConnectionHub struct {
	connOnHub    hub.ConnectionClient
	readTimeout  time.Duration
	writeTimeout time.Duration
	opts         options
	sync.Mutex
	sessionID *string
}

func New(connHub hub.ConnectionClient, opt ...ConnectionOptions) (conn *ConnectionHub, err error) {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	return &ConnectionHub{
		connOnHub:    connHub,
		readTimeout:  defTimeout,
		writeTimeout: defTimeout,
		opts:         opts,
	}, nil
}

func (c *ConnectionHub) Register() error {
	c.Lock()
	defer c.Unlock()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(c.opts.timeOut))
	defer cancel()
	res, err := c.connOnHub.Register(ctx, &hub.RegisterRequest{
		AgentId: c.opts.againtID,
	}, nil)

	// TODO better error handling will be here
	if err != nil {
		panic(err)
	}

	if res == nil {
		panic("Register response is empty")
	}
	switch res.GetError() {
	case hub.ConnectionErr_Error:
		return ErrAgentNotRegistered
	case hub.ConnectionErr_OK:
		session := res.GetSessionId()
		c.sessionID = &session
		return nil
	default:
		panic("not handled error after Register was called")

	}
}

func (c *ConnectionHub) isRegistered() bool {
	c.Lock()
	defer c.Unlock()
	return c.sessionID != nil
}

func (c *ConnectionHub) Read(b []byte) (n int, err error) {
	if !c.isRegistered() {
		return 0, io.ErrClosedPipe
	}
	c.Lock()
	defer c.Unlock()
	rSize := uint32(len(b))
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(c.opts.timeOut))
	defer cancel()
	res, err := c.connOnHub.Read(ctx, &hub.ReadRequest{
		Size:      rSize,
		SessionId: *c.sessionID,
	}, nil)

	// TODO better error handling will be here
	if err != nil {
		panic(err)
	}
	if res == nil {
		panic("Read response is empty")
	}

	switch res.GetErr() {
	case hub.ReadErr_EOF:
		return int(res.GetSize()), io.EOF
	case hub.ReadErr_UNEXPECTEDEOF:
		return 0, io.ErrUnexpectedEOF
	case hub.ReadErr_READ_CLOSEDPIPE:
		return 0, io.ErrClosedPipe
	}
	n = copy(b, res.GetData())

	return n, nil
}

func (c *ConnectionHub) Write(b []byte) (n int, err error) {
	if !c.isRegistered() {
		return 0, io.ErrClosedPipe
	}

	c.Lock()
	defer c.Unlock()
	wSize := len(b)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(c.opts.timeOut))
	defer cancel()
	res, err := c.connOnHub.Write(ctx, &hub.WriteRequest{
		Data:      b,
		SessionId: *c.sessionID,
	}, nil)

	// TODO better error handling will be here
	if err != nil {
		panic(err)
	}
	if res == nil {
		panic("Write response is empty")
	}

	n = int(res.GetSize())

	switch res.GetErr() {
	case hub.WriteErr_WRITE_OK:
		return
	case hub.WriteErr_SHORTWRITE:
		return n, io.ErrShortWrite
	case hub.WriteErr_WRITE_CLOSEDPIPE:
		return 0, io.ErrClosedPipe
	}

	if wSize > int(res.GetSize()) {
		return n, io.ErrShortWrite
	}
	return n, nil
}

func (c *ConnectionHub) Close() error {
	if !c.isRegistered() {
		return ErrAlreadyClosed
	}
	c.Lock()
	defer c.Unlock()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(c.opts.timeOut))
	defer cancel()
	res, err := c.connOnHub.Close(ctx, &hub.CloseRequest{SessionId: *c.sessionID})
	// TODO better error handling will be here
	if err != nil {
		panic(err)
	}

	if res == nil {
		panic("close response is empty")
	}

	switch res.GetErr() {
	case hub.ConnectionErr_OK:
		c.sessionID = nil
		return nil
	default:
		panic("not handled error after close was called")
	}

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
