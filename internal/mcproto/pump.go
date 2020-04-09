package mcproto

import (
	"context"
	"io"
	"net"
)

func (c *connectorImpl) pumpConnections(ctx context.Context, frontendConn, backendConn net.Conn) {
	//noinspection GoUnhandledErrorResult
	defer backendConn.Close()

	clientAddr := frontendConn.RemoteAddr()
	defer c.log.Debug().
		Str("client", clientAddr.String()).
		Str("backendConn", backendConn.RemoteAddr().String()).
		Msg("closing backend connection")

	errors := make(chan error, 2)
	go c.pumpFrames(backendConn, frontendConn, errors, "backend", "frontend", clientAddr)
	go c.pumpFrames(frontendConn, backendConn, errors, "frontend", "backend", clientAddr)

	select {
	case err := <-errors:
		if err != io.EOF {
			c.log.Error().
				Err(err).
				Str("client", clientAddr.String()).
				Str("backendConn", backendConn.RemoteAddr().String()).
				Msg("error on connection relay")
		}

	case <-ctx.Done():
		c.log.Debug().Msg("received context cancellation")
	}
}

func (c *connectorImpl) pumpFrames(incoming io.Reader, outgoing io.Writer, errors chan<- error, from, to string, clientAddr net.Addr) {
	amount, err := io.Copy(outgoing, incoming)
	c.log.Debug().
		Str("client", clientAddr.String()).
		Int64("amount", amount).
		Msgf("finished relay %s->%s", from, to)

	if err != nil {
		errors <- err
	} else {
		// successful io.Copy return nil error, not EOF… to simulate that to trigger outer handling
		errors <- io.EOF
	}
}
