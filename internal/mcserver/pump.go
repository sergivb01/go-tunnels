package mcserver

import (
	"context"
	"io"
	"net"
)

func (s *MCServer) pumpConnections(ctx context.Context, frontendConn, backendConn net.Conn) {
	clientAddr := frontendConn.RemoteAddr().String()
	defer s.log.Debug().
		Str("client", clientAddr).
		Str("backendConn", backendConn.RemoteAddr().String()).
		Msg("closing backend connection")

	errors := make(chan error, 2)
	go s.pumpFrames(backendConn, frontendConn, errors, "backend", "frontend", clientAddr)
	go s.pumpFrames(frontendConn, backendConn, errors, "frontend", "backend", clientAddr)

	select {
	case err := <-errors:
		if err != io.EOF {
			s.log.Error().Err(err).Str("client", clientAddr).
				Str("backend", backendConn.RemoteAddr().String()).
				Msg("on connection relay")
		}
	case <-ctx.Done():
		s.log.Debug().Msg("received context cancellation")
	}
}

func (s *MCServer) pumpFrames(incoming io.Reader, outgoing io.Writer, errors chan<- error, from, to string, clientAddr string) {
	amount, err := io.Copy(outgoing, incoming)
	if err != nil {
		errors <- err
	} else {
		// successful io.Copy return nil error, not EOF… to simulate that to trigger outer handling
		errors <- io.EOF
	}
	s.log.Debug().
		Str("client", clientAddr).
		Int64("amount", amount).
		Msgf("finished relay %s->%s", from, to)
}
