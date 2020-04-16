package mcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/juju/ratelimit"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"

	"github.com/sergivb01/mctunnel/internal/proto"
	"github.com/sergivb01/mctunnel/internal/proto/packet"
)

const handshakeTimeout = 1 * time.Second

var noDeadline time.Time

type MCServer struct {
	log         zerolog.Logger
	packetCoder proto.PacketCodec
	c           *cache.Cache
}

func NewConnector() *MCServer {
	return &MCServer{
		log:         zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().Level(zerolog.InfoLevel),
		packetCoder: proto.NewPacketCodec(),
		c:           cache.New(time.Minute*5, time.Minute*10),
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

	log := s.log.With().Str("client", frontendConn.RemoteAddr().String()).Logger()

	if err := frontendConn.SetNoDelay(true); err != nil {
		log.Error().Err(err).Msg("error setting TCPNoDelay to frontendConn")
	}

	if err := frontendConn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		log.Error().Err(err).Msg("failed to set read deadline")
		return
	}

	packetID, err := s.packetCoder.ReadPacket(frontendConn)
	if err != nil {
		log.Error().Err(err).Msg("error reading packetID")
		return
	}

	if packetID != packet.HandshakeId {
		log.Error().Int("packetID", packetID).Msg("received unknown first packet")
		return
	}

	h := &packet.Handshake{}
	if err := h.Decode(frontendConn); err != nil {
		log.Error().Err(err).Msg("error reading handshake")
		return
	}

	// TODO: if h.Status == 1, return custom StatusResponse, otherwise read LoginStartPacket (https://wiki.vg/Protocol#Login_Start)
	if err = frontendConn.SetReadDeadline(noDeadline); err != nil {
		log.Error().Err(err).Msg("failed to clear read deadline")
		return
	}

	s.findAndConnectBackend(ctx, frontendConn, h, t)
}

func (s *MCServer) findAndConnectBackend(ctx context.Context, frontendConn *net.TCPConn, h *packet.Handshake, t time.Time) {
	log := s.log.With().Str("client", frontendConn.RemoteAddr().String()).Str("handshakeAddres", h.ServerAddress).Uint16("handshakePort", h.ServerPort).Logger()

	login := &packet.LoginStart{}
	if h.State == 2 {
		packetID, err := s.packetCoder.ReadPacket(frontendConn)
		if err != nil {
			log.Error().Err(err).Msg("error reading packetID")
			return
		}

		if packetID != packet.HandshakeId {
			log.Error().Int("packetID", packetID).Msg("received unknown second packet")
			return
		}

		if err := login.Decode(frontendConn); err != nil {
			log.Error().Err(err).Msg("error decoding LoginStart")
			return
		}

		log = log.With().Str("playerName", login.Name).Logger()
		log.Debug().Str("playerName", login.Name).Msg("read playerName from LoginStart")
	}

	host, port := s.ResolveServerAddress(h.ServerAddress)
	log.Info().Str("hostPort", host+":"+port).Msg("found backend for connection")

	addr, err := net.ResolveTCPAddr("tcp", host+":"+port)
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
	if err := s.packetCoder.WritePacket(remote, h); err != nil {
		log.Error().Err(err).Msg("failed to relay handshake!")
		return
	}

	if h.State == 2 {
		if err := s.packetCoder.WritePacket(remote, login); err != nil {
			log.Error().Err(err).Msg("failed to relay login!")
			return
		}
	}

	log.Info().Str("took", time.Since(t).String()).Msg("pipe with remote started")
	s.pumpConnections(ctx, frontendConn, remote)
	log.Info().Msg("piped with remote closed, connection closed")
}
