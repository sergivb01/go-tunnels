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
	"github.com/sergivb01/mctunnel/internal/proto/packet"
)

const handshakeTimeout = 1 * time.Second

var noDeadline time.Time

type MCServer struct {
	log         zerolog.Logger
	packetCoder proto.PacketCodec
}

func NewConnector() *MCServer {
	return &MCServer{
		log:         zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().Level(zerolog.InfoLevel),
		packetCoder: proto.NewPacketCodec(),
	}
}

func (s *MCServer) Start(ctx context.Context, listenAddress string, connRateLimit int) error {
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
				s.log.Error().Err(err).Str("remoteAddr", conn.RemoteAddr().String()).
					Msg("error accepting connection")
			} else {
				go s.handleConnection(ctx, conn, time.Now())
			}
		}
	}
}

func (s *MCServer) handleConnection(ctx context.Context, frontendConn *net.TCPConn, t time.Time) {
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

	packetID, err := s.packetCoder.DecodePacket(frontendConn)
	if err != nil {
		cLog.Error().Err(err).Msg("error reading packetID")
		return
	}

	if packetID != packet.HandshakeId {
		cLog.Error().Int("packetID", packetID).Msg("received first unknown packet")
		return
	}

	h := &packet.Handshake{}
	if err := h.Decode(frontendConn); err != nil {
		cLog.Error().Err(err).Msg("error reading handshake")
		return
	}

	// TODO: if h.Status == 1, return custom StatusResponse
	if err = frontendConn.SetReadDeadline(noDeadline); err != nil {
		cLog.Error().Err(err).Msg("failed to clear read deadline")
		return
	}

	s.findAndConnectBackend(ctx, frontendConn, h, t)
}

func (s *MCServer) findAndConnectBackend(ctx context.Context, frontendConn *net.TCPConn, h *packet.Handshake, t time.Time) {
	log := s.log.With().Str("client", frontendConn.RemoteAddr().String()).Str("handshakeAddres", h.ServerAddress).Uint16("handshakePort", h.ServerPort).Logger()

	host, port := ResolveServerAddress(h.ServerAddress)
	log.Info().Str("host", host).Int("port", port).Msg("found backend for connection")

	addr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Error().Err(err).Msg("error resolving tcp address")
		return
	}

	remote, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Error().Err(err).Msg("unable to connect to backend")
		return
	}
	defer remote.Close()

	if err := remote.SetNoDelay(true); err != nil {
		log.Error().Err(err).Msg("error setting TCPNoDelay to remote")
	}

	h.ServerAddress = host
	if err := s.packetCoder.EncodePacket(remote, h); err != nil {
		log.Error().Err(err).Msg("failed to relay handshake!")
		return
	}

	log.Info().Str("took", time.Since(t).String()).Msg("pipe with remote started")
	s.pumpConnections(ctx, frontendConn, remote)
	log.Info().Msg("piped with remote closed, connection closed")
}
