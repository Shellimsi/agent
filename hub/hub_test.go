package hub

import (
	"bytes"
	"io"
	"math/rand"
	"net"
	"testing"

	hub "github.com/shellimsi/proto/hub"
	"github.com/stretchr/testify/assert"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

var _ net.Conn = &ConnectionHub{}
var _ hub.ConnectionClient = &ReadMock{}

type ReadMock struct {
	buffer *bytes.Buffer
}

func (c *ReadMock) Read(ctx context.Context, in *hub.ReadRequest, opts ...grpc.CallOption) (res *hub.ReadResponse, err error) {
	size := in.GetSize()
	res = new(hub.ReadResponse)
	res.TotalSize = uint32(c.buffer.Cap())
	if int(size) >= c.buffer.Len() {
		res.Err = hub.ConnectionErr_EOF
		return
	}

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
	return
}

func NewReadMock() *ReadMock {
	randByt := make([]byte, 1024)
	rand.Read(randByt)
	return &ReadMock{
		buffer: bytes.NewBuffer(randByt),
	}
}

func TestFirstRead(t *testing.T) {
	rMock := NewReadMock()
	conn, err := New(rMock)

	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 512, n, "The two words should be the same.")

	buf2 := make([]byte, 512)
	n2, err := conn.Read(buf2)
	if err != io.EOF {
		t.Fatal(err)
	}

	assert.Equal(t, 512, n2, "The two words should be the same.")

}
