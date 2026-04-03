package rtmp

import (
	"fmt"
	"io"
	"log"
	"net"

	rtmp "github.com/yutopp/go-rtmp"
)

// Server wraps the RTMP server
type Server struct {
	addr    string
	handler *StreamHandler
	server  *rtmp.Server
	ln      net.Listener
}

// NewServer creates a new RTMP server
func NewServer(addr string, handler *StreamHandler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}

// Start starts the RTMP server
func (s *Server) Start() error {
	// Create TCP listener
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	s.ln = ln

	log.Printf("[RTMP] Server listening on %s", s.addr)
	log.Printf("[RTMP] Ready to accept RTMP streams")
	log.Printf("[RTMP] Example OBS settings:")
	log.Printf("[RTMP]   Server: rtmp://<your-ip>%s/live", s.addr)
	log.Printf("[RTMP]   Stream Key: <your_stream_key>")

	// Create RTMP server with configuration
	s.server = rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[RTMP] PANIC in OnConnect callback: %v", r)
				}
			}()

			log.Printf("[RTMP] Incoming connection from %s", conn.RemoteAddr())

			if s.handler == nil {
				log.Printf("[RTMP] ERROR: Handler is nil!")
				return conn, nil
			}

			// Create a new ConnectionHandler for this specific connection
			// This ensures each connection tracks its own stream independently
			connHandler := s.handler.NewConnectionHandler()
			log.Printf("[RTMP] Created new connection handler for %s", conn.RemoteAddr())

			// Return connection config with per-connection handler
			config := &rtmp.ConnConfig{
				Handler: connHandler,

				// Control state configuration
				ControlState: rtmp.StreamControlStateConfig{
					DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
				},
			}

			return conn, config
		},
	})

	// Serve connections
	if err := s.server.Serve(ln); err != nil {
		return fmt.Errorf("RTMP server error: %w", err)
	}

	return nil
}

// Stop stops the RTMP server gracefully
func (s *Server) Stop() error {
	log.Printf("[RTMP] Shutting down server...")

	if s.ln != nil {
		if err := s.ln.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	log.Printf("[RTMP] Server stopped")
	return nil
}

// GetActiveStreamsCount returns the number of active streams
func (s *Server) GetActiveStreamsCount() int {
	return len(s.handler.GetActiveStreams())
}
