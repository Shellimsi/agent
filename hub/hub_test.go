package hub

import (
	"io"
	"net"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
	hub "github.com/shellimsi/proto/hub"
	"github.com/stretchr/testify/assert"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

var _ net.Conn = &ConnectionHub{}
var _ hub.ConnectionClient = &ReadMock{}

type ReadMock struct {
	buffer io.Reader
}

func (c *ReadMock) reset() {
	c.buffer = strings.NewReader("")
}

func (c *ReadMock) Close(ctx context.Context, in *hub.CloseRequest, opts ...grpc.CallOption) (res *hub.CloseResponse, err error) {
	return &hub.CloseResponse{
		Err: hub.ConnectionErr_OK,
	}, nil
}

func (c *ReadMock) Register(ctx context.Context, in *hub.RegisterRequest, opts ...grpc.CallOption) (res *hub.RegisterResponse, err error) {
	return &hub.RegisterResponse{
		AgentId:   in.GetAgentId(),
		SessionId: uuid.NewV4().String(),
	}, nil
}

func (c *ReadMock) Write(ctx context.Context, in *hub.WriteRequest, opts ...grpc.CallOption) (res *hub.WriteResponse, err error) {
	return nil, nil
}

func (c *ReadMock) Read(ctx context.Context, in *hub.ReadRequest, opts ...grpc.CallOption) (res *hub.ReadResponse, err error) {
	size := in.GetSize()
	res = new(hub.ReadResponse)

	buff := make([]byte, size)
	n, err := c.buffer.Read(buff)
	res.Size = uint32(n)
	res.Data = buff
	if err != nil {
		switch err {
		case io.EOF:
			res.Err = hub.ReadErr_EOF
			return res, nil
		default:
			panic("not handled")
		}
	}
	res.Data = buff
	res.Err = hub.ReadErr_READ_OK
	return
}

func NewReadMock() *ReadMock {
	return &ReadMock{
		buffer: strings.NewReader("hello world"),
	}
}

func TestFirstRead(t *testing.T) {
	rMock := NewReadMock()
	conn, err := New(rMock)
	if err != nil {
		t.Fatal(err)
	}
	err = conn.Register()
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 3)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(buf), n, "should be the same.")
	assert.Equal(t, string(buf), "hel", "should be the same.")

	buf2 := make([]byte, 8)
	n2, err := conn.Read(buf2)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(buf2), n2, "should be the same.")
	assert.Equal(t, string(buf2), "lo world", "should be the same.")

	buf3 := make([]byte, 10)
	n3, err := conn.Read(buf3)
	if err != io.EOF {
		t.Fatal(err)
	}
	assert.Equal(t, 0, n3, "should be the same.")

	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}
}
