package hub

import (
	"io"
	"net"
	"strings"
	"testing"

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
			res.Err = hub.ConnectionErr_EOF
			return res, nil
		default:
			panic("not handled")
		}
	}

	res.Data = buff

	res.Err = hub.ConnectionErr_OK
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
	buf := make([]byte, 3)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(buf), n, "The two words should be the same.")
	assert.Equal(t, string(buf), "hel", "The two words should be the same.")

	buf2 := make([]byte, 8)
	n2, err := conn.Read(buf2)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(buf2), n2, "The two words should be the same.")
	assert.Equal(t, string(buf2), "lo world", "The two words should be the same.")

	buf3 := make([]byte, 10)
	n3, err := conn.Read(buf3)
	if err != io.EOF {
		t.Fatal(err)
	}
	assert.Equal(t, 0, n3, "The two words should be the same.")
}
