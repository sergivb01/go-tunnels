package mcserver

import (
	"context"
	"io"
	"net"
)

func (s *MCServer) pumpConnections(ctx context.Context, conn, remote *net.TCPConn) {
	clientAddr := conn.RemoteAddr().String()
	defer s.log.Debug().Str("client", clientAddr).Str("remote", remote.RemoteAddr().String()).
		Msg("closing backend connection")

	errors := make(chan error, 2)
	go s.pumpFrames(remote, conn, errors, "backend", "client", clientAddr)
	go s.pumpFrames(conn, remote, errors, "client", "backend", clientAddr)

	select {
	case err := <-errors:
		if err != io.EOF {
			s.log.Error().Err(err).Str("client", clientAddr).
				Str("backend", remote.RemoteAddr().String()).Msg("on connection relay")
		}
	case <-ctx.Done():
		s.log.Debug().Msg("received context cancellation")
	}
}

func (s *MCServer) pumpFrames(incoming, outgoing *net.TCPConn, errors chan<- error, from, to, clientAddr string) {
	amount, err := io.Copy(outgoing, incoming)
	if err != nil {
		errors <- err
	} else {
		// successful io.Copy return nil error, not EOF to simulate that to trigger outer handling
		errors <- io.EOF
	}
	s.log.Info().Str("client", clientAddr).Int64("amount", amount).
		Msgf("finished relay %s->%s", from, to)
}
