package control

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"
	"sync"
)

type HandlerFunc func(Command) string

type Server struct {
	address string
	handler HandlerFunc

	listener net.Listener
	wg       sync.WaitGroup
}

// NewServer creates a Server Struct
func NewServer(address string, handler HandlerFunc) *Server {
	return &Server{
		address: address,
		handler: handler,
	}
}

// Starts the server
func (s *Server) Start() error {

	// Opens port
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	// Saves listener
	s.listener = l

	// wait until connection
	for {
		conn, err := l.Accept() // New connection
		// Checks
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			continue
		}

		s.wg.Add(1)

		// Server interaction
		go s.handleConn(conn)
	}
}

// Stop strops the server
func (s *Server) Stop(ctx context.Context) error {
	if s.listener == nil {
		return nil
	}

	_ = s.listener.Close()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// On each connection
func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	// Welcome message
	writeLine(conn, "Welcome to CHIMERA control server. Type 'h' for help.")

	// Reader
	scanner := bufio.NewScanner(conn)

	// For new line
	for scanner.Scan() {

		// Parse
		line := scanner.Text()
		cmd := ParseCommand(line)

		if len(cmd) == 0 {
			writeLine(conn, "EMPTY")
			continue
		}

		// Allow clients to close connection
		switch strings.ToLower(cmd[0]) {
		case "quit", "exit", "bye":
			writeLine(conn, "BYE")
			return
		}

		// Compute answer
		resp := s.handler(cmd)

		// Print output
		writeLine(conn, resp)
	}

	if err := scanner.Err(); err != nil {
		writeLine(conn, "ERROR: "+err.Error())
	}
}

// write line helper
func writeLine(conn net.Conn, msg string) {
	_, _ = conn.Write([]byte(msg + "\n"))
}
