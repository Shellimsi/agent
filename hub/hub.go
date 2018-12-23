package hub

import (
	"context"
	"fmt"
	"time"

	hub "github.com/shellimsi/proto/hub"
	"google.golang.org/grpc"
)

func NewConn(host string, port uint) (*grpc.ClientConn, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type Client struct {
	Host string
	Port uint
}

func NewClient(host string, port uint) *Client {
	return &Client{
		Host: host,
		Port: port,
	}
}

func (c *Client) Register(req *hub.RegisterRequest) (*hub.RegisterResponse, error) {
	conn, err := NewConn(c.Host, c.Port)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	h := hub.NewTerminalClient(conn)
	return h.Register(ctx, req)
}
