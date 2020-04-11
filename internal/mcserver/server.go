package mcserver

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/juju/ratelimit"
	"github.com/rs/zerolog"

	"github.com/sergivb01/mctunnel/internal/proto"
)

const handshakeTimeout = 3 * time.Second

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
	addr, err := net.ResolveTCPAddr("tcp", listenAddress)
	if err != nil {
		return fmt.Errorf("error resolving local adderss: %w", err)
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		s.log.Error().Err(err).Msg("unable to start listening")
		return err
	}
	s.log.Info().Str("listenAddress", listenAddress).Msg("listening for MC client connections")

	return s.acceptConnections(ctx, ln, connRateLimit)
}

func (s *MCServer) acceptConnections(ctx context.Context, ln *net.TCPListener, connRateLimit int) error {
	bucket := ratelimit.NewBucketWithRate(float64(connRateLimit), int64(connRateLimit*2))

	for {
		select {
		case <-ctx.Done():
			return ln.Close()

		case <-time.After(bucket.Take(1)):
			conn, err := ln.AcceptTCP()
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

func (s *MCServer) handleConnection(ctx context.Context, frontendConn *net.TCPConn) {
	// //noinspection GoUnhandledErrorResult
	defer frontendConn.Close()

	cLog := s.log.With().Str("client", frontendConn.RemoteAddr().String()).Logger()

	if err := frontendConn.SetNoDelay(true); err != nil {
		cLog.Error().Err(err).Msg("error setting TCPNoDelay to frontendConn")
	}

	if err := frontendConn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		cLog.Error().Err(err).Msg("failed to set read deadline")
		return
	}

	mcConn := &proto.MCConn{TCPConn: frontendConn}

	packetID, reader, err := proto.PacketReader(mcConn)
	if err != nil {
		cLog.Error().Err(err).Msg("error creating packet reader")
		return
	}

	if packetID != 0x00 {
		cLog.Info().Uint64("packetID", packetID).Msg("received first unknown packet")
		return
	}

	h, err := proto.ReadHandshake(reader)
	if err != nil {
		cLog.Error().Err(err).Msg("error reading handshake")
		return
	}

	// TODO: if h.Status == 1, return custom StatusResponse

	if err = frontendConn.SetReadDeadline(noDeadline); err != nil {
		cLog.Error().Err(err).Msg("failed to clear read deadline")
		return
	}

	s.findAndConnectBackend(ctx, frontendConn, h)
}

func (s *MCServer) findAndConnectBackend(ctx context.Context, frontendConn *net.TCPConn, h *proto.Handshake) {
	cLog := s.log.With().Str("client", frontendConn.RemoteAddr().String()).Str("handshakeAddres", h.Address).Uint16("handshakePort", h.Port).Logger()
	cLog.Info().Msg("connecting to backend")

	host, port := ResolveServerAddress(h.Address)
	cLog.Info().Str("host", host).Int("port", port).Msg("found backend for connection")

	addr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		cLog.Error().Err(err).Msg("error resolving tcp address")
		return
	}

	remote, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		cLog.Error().Err(err).Msg("unable to connect to backend")
		return
	}
	defer remote.Close()

	if err := remote.SetNoDelay(true); err != nil {
		cLog.Error().Err(err).Msg("error setting TCPNoDelay to remote")
	}

	if err := h.Write(remote, host); err != nil {
		cLog.Error().Err(err).Msg("failed to relay handshake!")
		return
	}

	s.pumpConnections(ctx, frontendConn, remote)
}
