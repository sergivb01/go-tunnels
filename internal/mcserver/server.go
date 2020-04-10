package mcserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/juju/ratelimit"
	"github.com/rs/zerolog"

	"github.com/sergivb01/mctunnel/internal/proto"
)

const handshakeTimeout = 5 * time.Second

var noDeadline time.Time

func NewConnector() *MCServer {
	return &MCServer{
		log: zerolog.New(zerolog.NewConsoleWriter()).
			With().Timestamp().
			Logger().Level(zerolog.DebugLevel),
	}
}

type MCServer struct {
	log zerolog.Logger
}

func (s *MCServer) StartAcceptingConnections(ctx context.Context, listenAddress string, connRateLimit int) error {
	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		s.log.Error().Err(err).Msg("unable to start listening")
		return err
	}
	s.log.Info().Str("listenAddress", listenAddress).Msg("listening for MC client connections")

	return s.acceptConnections(ctx, ln, connRateLimit)
}

func (s *MCServer) acceptConnections(ctx context.Context, ln net.Listener, connRateLimit int) error {
	bucket := ratelimit.NewBucketWithRate(float64(connRateLimit), int64(connRateLimit*2))

	for {
		select {
		case <-ctx.Done():
			return ln.Close()

		case <-time.After(bucket.Take(1)):
			conn, err := ln.Accept()
			if err != nil {
				s.log.Error().
					Err(err).
					Str("remoteAddr", conn.RemoteAddr().String()).
					Msg("error accepting connection")
			} else {
				go s.handleConnection(ctx, conn)
			}
		}
	}
}

func (s *MCServer) handleConnection(ctx context.Context, frontendConn net.Conn) {
	//noinspection GoUnhandledErrorResult
	defer frontendConn.Close()

	clientAddr := frontendConn.RemoteAddr()
	cLog := s.log.With().Str("client", clientAddr.String()).Logger()
	cLog.Info().Msg("got connection")
	defer cLog.Info().Msg("closing connection")

	if err := frontendConn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		cLog.Error().Err(err).Msg("failed to set read deadline")
		return
	}

	player := proto.NewPlayer(frontendConn)
	player.ReadPacket()

	s.findAndConnectBackend(ctx, frontendConn, clientAddr, buf, serverAddress, customHandshake)
}

func (s *MCServer) findAndConnectBackend(ctx context.Context, frontendConn net.Conn,
	clientAddr net.Addr, preReadContent io.Reader, serverAddress string) {
	// backendHostPort, resolvedHost := Routes.FindBackendForServerAddress(serverAddress)
	// host, port, err := ExtractHostPort(serverAddress)
	host, port := "mc.hypixel.net", 25565
	var err error
	if err != nil {
		s.log.Error().Err(err).Str("client", clientAddr.String()).Str("server", serverAddress).
			Msg("could not find backend")
		return
	}
	backendHostPort := fmt.Sprintf("%s:%d", host, port)
	cLog := s.log.With().Str("client", clientAddr.String()).Str("backendHostPort", backendHostPort).Str("server", serverAddress).Logger()

	cLog.Info().Msg("connecting to backend")

	backendConn, err := net.Dial("tcp", backendHostPort)
	if err != nil {
		cLog.Error().Err(err).Msg("unable to connect to backend")
		return
	}

	amount, err := io.Copy(backendConn, preReadContent)
	if err != nil {
		cLog.Error().Err(err).Msg("failed to write handshake to backend")
		return
	}
	cLog.Debug().Int64("amout", amount).Msg("relayed handshake to backend")

	if err = frontendConn.SetReadDeadline(noDeadline); err != nil {
		cLog.Error().Err(err).Msg("failed to clear read deadline")
		return
	}

	s.pumpConnections(ctx, frontendConn, backendConn)
}
