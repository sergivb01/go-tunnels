package mcproto

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	cLog := c.log.With().Str("client", clientAddr.String()).Logger()
	cLog.Info().Msg("got connection")
	defer cLog.Info().Msg("closing connection")

	if err := frontendConn.SetNoDelay(true); err != nil {
		cLog.Error().Err(err).Msg("failed to set TCPNoDelay")
	}

	inspectionBuffer := new(bytes.Buffer)
	inspectionReader := io.TeeReader(frontendConn, inspectionBuffer)

	if err := frontendConn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		cLog.Error().Err(err).Msg("failed to set read deadline")
		return
	}

	packet, err := ReadPacket(inspectionReader, c.state)
	if err != nil {
		cLog.Error().Err(err).Msg("failed to read packet")
		return
	}

	cLog.Debug().Int("length", packet.Length).Int("packetID", packet.PacketID).Msg("received packet")

	var (
		serverAddress string
	)
	customHandshake := &Handshake{
		ProtocolVersion: -1,
	}

	switch packet.PacketID {
	case PacketIdHandshake:
		handshake, err := ReadHandshake(packet.Data)
		if err != nil {
			cLog.Error().Err(err).Msg("failed to read handshake")
			return
		}
		cLog.Debug().Interface("handshake", handshake).Msg("received handshake")
		serverAddress = handshake.ServerAddress
		customHandshake = handshake
	case PacketIdLegacyServerListPing:
		handshake, ok := packet.Data.(*LegacyServerListPing)
		if !ok {
			cLog.Error().Err(err).Interface("packet", packet).Msg("unexpected data type for PacketIdLegacyServerListPing")
			return
		}
		cLog.Debug().Interface("handshake", handshake).Msg("received legacy server list ping")
		serverAddress = handshake.ServerAddress
	default:
		cLog.Error().Interface("packet", packet).Int("packetID", packet.PacketID).Msg("unexpected content")
		return
	}

	c.findAndConnectBackend(ctx, frontendConn, clientAddr, inspectionBuffer, serverAddress, customHandshake)
}

func (c *connectorImpl) findAndConnectBackend(ctx context.Context, frontendConn net.Conn,
	clientAddr net.Addr, preReadContent io.Reader, serverAddress string, h *Handshake) {
	// backendHostPort, resolvedHost := Routes.FindBackendForServerAddress(serverAddress)
	host, port, err := ExtractHostPort(serverAddress)
	if err != nil {
		c.log.Error().Err(err).Str("client", clientAddr.String()).Str("server", serverAddress).
			Msg("could not find backend")
		return
	}
	backendHostPort := fmt.Sprintf("%s:%d", host, port)
	cLog := c.log.With().Str("client", clientAddr.String()).Str("backendHostPort", backendHostPort).Str("server", serverAddress).Logger()

	cLog.Info().Msg("connecting to backend")

	backendConn, err := net.Dial("tcp", backendHostPort)
	if err != nil {
		cLog.Error().Err(err).Msg("unable to connect to backend")
		return
	}

	if h.ProtocolVersion != -1 && h.NextState == 2 {
		b, err := h.EncodePacket(host)
		if err != nil {
			cLog.Error().Err(err).Msg("failed to enconde custom handshake")
			return
		}

		amount, err := backendConn.Write(b)
		if err != nil {
			cLog.Error().Err(err).Msg("1234failed to write handshake to backend")
			return
		}
		cLog.Debug().Int("amout", amount).Msg("1234relayed handshake to backend")
	} else {
		amount, err := io.Copy(backendConn, preReadContent)
		if err != nil {
			cLog.Error().Err(err).Msg("failed to write handshake to backend")
			return
		}
		cLog.Debug().Int64("amout", amount).Msg("relayed handshake to backend")
	}

	if err = frontendConn.SetReadDeadline(noDeadline); err != nil {
		cLog.Error().Err(err).Msg("failed to clear read deadline")
		return
	}

	c.pumpConnections(ctx, frontendConn, backendConn)
}
