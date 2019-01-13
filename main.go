package main

import (
	"fmt"
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/shellimsi/proto/hub"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
	"github.com/satori/go.uuid"
)

func newShell(conn net.Conn) error {
	// Create arbitrary command.
	c := exec.Command("bash")

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	go func() { _, _ = io.Copy(ptmx, conn) }()
	_, _ = io.Copy(conn, ptmx)

	return nil
}

func main() {
	

	// Set up a connection to the server.
	conngRPC, err := grpc.Dial("localhost:8090", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conngRPC.Close()
	client := hub.NewTerminalClient(conngRPC )
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := client.Register(ctx, &hub.RegisterRequest{
		TerminalID: uuid.NewV4().String(),
		AgentID: "agentID-1",
		Address : "192.168.1.1",
		WSize : nil,
	})

	if err != nil {
		panic(err)
	}

	log.Printf("Connecting to %s:%d", res.Host, res.Port)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", res.Host, res.Port))
	if err != nil {
		panic(err)
	}
	if err := newShell(conn); err != nil {
		log.Fatal(err)
	}
}
