package mcproto

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/juju/ratelimit"
	"github.com/rs/zerolog"
)

const handshakeTimeout = 5 * time.Second

var noDeadline time.Time

type Connector interface {
	StartAcceptingConnections(ctx context.Context, listenAddress string, connRateLimit int) error
	EncodeDecode()
}

func NewConnector() Connector {
	return &connectorImpl{
		log: zerolog.New(zerolog.NewConsoleWriter()).
			With().Timestamp().
			Logger().Level(zerolog.DebugLevel),
	}
}

type connectorImpl struct {
	state State
	log   zerolog.Logger
}

func (c *connectorImpl) StartAcceptingConnections(ctx context.Context, listenAddress string, connRateLimit int) error {
	addr, err := net.ResolveTCPAddr("tcp", ":25565")
	if err != nil {
		return fmt.Errorf("failed to resolve TCP listen address: %w", err)
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		c.log.Error().Err(err).Msg("unable to start listening")
		return err
	}
	c.log.Info().Str("listenAddress", listenAddress).Msg("listening for MC client connections")

	return c.acceptConnections(ctx, ln, connRateLimit)
}

func (c *connectorImpl) acceptConnections(ctx context.Context, ln *net.TCPListener, connRateLimit int) error {
	bucket := ratelimit.NewBucketWithRate(float64(connRateLimit), int64(connRateLimit*2))

	for {
		select {
		case <-ctx.Done():
			return ln.Close()

		case <-time.After(bucket.Take(1)):
			conn, err := ln.AcceptTCP()
			if err != nil {
				c.log.Error().
					Err(err).
					Str("remoteAddr", conn.RemoteAddr().String()).
					Msg("error accepting connection")
			} else {
				go c.HandleConnection(ctx, conn)
			}
		}
	}
}

func (c *connectorImpl) HandleConnection(ctx context.Context, frontendConn *net.TCPConn) {
	//noinspection GoUnhandledErrorResult
	defer frontendConn.Close()

	clientAddr := frontendConn.RemoteAddr()
	c.log.Info().Str("client", clientAddr.String()).Msg("got connection")

	if err := frontendConn.SetNoDelay(true); err != nil {
		c.log.Error().Err(err).Str("client", clientAddr.String()).Msg("failed to set TCPNoDelay")
	}

	defer c.log.Info().Str("client", clientAddr.String()).Msg("closing connection")

	inspectionBuffer := new(bytes.Buffer)
	inspectionReader := io.TeeReader(frontendConn, inspectionBuffer)

	if err := frontendConn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		c.log.Error().
			Err(err).
			Str("client", clientAddr.String()).
			Msg("failed to set read deadline")
		return
	}

	packet, err := ReadPacket(inspectionReader, c.state)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("client", clientAddr.String()).
			Msg("failed to read packet")
		return
	}

	c.log.Debug().
		Str("client", clientAddr.String()).
		Int("length", packet.Length).
		Int("packetID", packet.PacketID).
		Msg("received packet")

	var serverAddress string
	customHandshake := &Handshake{ProtocolVersion: -1}

	switch packet.PacketID {
	case PacketIdHandshake:
		handshake, err := ReadHandshake(packet.Data)
		if err != nil {
			c.log.Error().
				Err(err).
				Str("client", clientAddr.String()).
				Msg("failed to read handshake")
			return
		}
		c.log.Debug().
			Str("client", clientAddr.String()).
			Interface("handshake", handshake).
			Msg("received handshake")
		serverAddress = handshake.ServerAddress
		customHandshake = handshake
	case PacketIdLegacyServerListPing:
		handshake, ok := packet.Data.(*LegacyServerListPing)
		if !ok {
			c.log.Error().
				Err(err).
				Str("client", clientAddr.String()).
				Interface("packet", packet).
				Msg("unexpected data type for PacketIdLegacyServerListPing")
			return
		}
		c.log.Debug().
			Str("client", clientAddr.String()).
			Interface("handshake", handshake).
			Msg("received legacy server list ping")
		serverAddress = handshake.ServerAddress
	default:
		c.log.Error().
			Str("client", clientAddr.String()).
			Interface("packet", packet).
			Int("packetID", packet.PacketID).
			Msg("unexpected content")
		return
	}

	c.findAndConnectBackend(ctx, frontendConn, clientAddr, inspectionBuffer, serverAddress, customHandshake)
}

func (c *connectorImpl) findAndConnectBackend(ctx context.Context, frontendConn net.Conn,
	clientAddr net.Addr, preReadContent io.Reader, serverAddress string, h *Handshake) {

	// backendHostPort, resolvedHost := Routes.FindBackendForServerAddress(serverAddress)
	backendHostPort, err := ExtractHostPort(serverAddress)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("client", clientAddr.String()).
			Str("server", serverAddress).
			Msg("could not find backend")
		return
	}

	c.log.Info().
		Str("client", clientAddr.String()).
		Str("server", serverAddress).
		Str("backendHostPort", backendHostPort).
		Msg("connecting to backend")

	backendConn, err := net.Dial("tcp", backendHostPort)
	if err != nil {
		c.log.Error().Err(err).
			Str("client", clientAddr.String()).
			Str("serverAddress", serverAddress).
			Str("backendHostPort", backendHostPort).
			Msg("unable to connect to backend")
		return
	}

	if h.ProtocolVersion != -1 && h.NextState == 2{
		b, err := h.EncodePacket("na.lunar.gg")
		if err != nil {
			c.log.Error().
				Err(err).
				Str("client", clientAddr.String()).
				Str("backendHostPort", backendHostPort).
				Msg("failed to enconde custom handshake")
		}
		amount, err := backendConn.Write(b)
		if err != nil {
			c.log.Error().
				Err(err).
				Str("client", clientAddr.String()).
				Str("backendHostPort", backendHostPort).
				Msg("1234failed to write handshake to backend")
			return
		}
		c.log.Debug().
			Str("client", clientAddr.String()).
			Str("backendHostPort", backendHostPort).
			Int("amout", amount).
			Msg("1234relayed handshake to backend")
	} else {
		amount, err := io.Copy(backendConn, preReadContent)
		if err != nil {
			c.log.Error().
				Err(err).
				Str("client", clientAddr.String()).
				Str("backendHostPort", backendHostPort).
				Msg("failed to write handshake to backend")
			return
		}
		c.log.Debug().
			Str("client", clientAddr.String()).
			Str("backendHostPort", backendHostPort).
			Int64("amout", amount).
			Msg("relayed handshake to backend")
	}

	if err = frontendConn.SetReadDeadline(noDeadline); err != nil {
		c.log.Error().
			Err(err).
			Str("client", clientAddr.String()).
			Str("backendHostPort", backendHostPort).
			Msg("failed to clear read deadline")
		return
	}

	c.pumpConnections(ctx, frontendConn, backendConn)
}

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
			log.Printf("error in conn relay: %s", err)
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
