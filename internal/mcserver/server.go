package mcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/juju/ratelimit"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"

	"github.com/minebreach/go-tunnels/internal/proto"
	"github.com/minebreach/go-tunnels/internal/proto/packet"
)

const handshakeTimeout = 1 * time.Second

var noDeadline time.Time

// MCServer defines a Minecraft relay server
type MCServer struct {
	log         zerolog.Logger
	packetCoder proto.PacketCodec
	c           *cache.Cache
}

// NewConnector creates a new MCServer
func NewConnector() *MCServer {
	return &MCServer{
		log:         zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().Level(zerolog.InfoLevel),
		packetCoder: proto.NewPacketCodec(),
		c:           cache.New(time.Minute*5, time.Minute*10),
	}
}

// Start starts listening for new connections
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
	defer func() {
		if err := frontendConn.Close(); err != nil {
			s.log.Error().Err(err).Str("client", frontendConn.RemoteAddr().String()).Msg("error closing frontend connection")
		}
	}()
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

	if packetID != packet.HandshakeID {
		log.Error().Int("packetID", packetID).Msg("received unknown first packet")
		return
	}

	h := &packet.Handshake{}
	if err := h.Decode(frontendConn); err != nil {
		log.Error().Err(err).Msg("error reading handshake")
		return
	}

	if h.State == 1 {
		// TODO: save this packet in const and send it to frontendConn on error
		if err := s.packetCoder.WritePacket(frontendConn, &packet.ServerStatus{
			ServerName: "Minebreach",
			Protocol:   h.ProtocolVersion,
			Motd:       "§3§lMineBreach Tunnels\n§cError - Unknown hostname",
			Favicon:    "",
		}); err != nil {
			log.Error().Err(err).Msg("error sending custom ServerListPing response")
			return
		}
		return
	}

	if err = frontendConn.SetReadDeadline(noDeadline); err != nil {
		log.Error().Err(err).Msg("failed to clear read deadline")
		return
	}

	s.findAndConnectBackend(ctx, frontendConn, h, t)
}

func (s *MCServer) findAndConnectBackend(ctx context.Context, frontendConn *net.TCPConn, h *packet.Handshake, t time.Time) {
	log := s.log.With().Str("client", frontendConn.RemoteAddr().String()).Str("handshakeAddres", h.ServerAddress).Uint16("handshakePort", h.ServerPort).Logger()

	login := &packet.LoginStart{}
	// TODO: cleanup logic
	if h.State == 2 {
		packetID, err := s.packetCoder.ReadPacket(frontendConn)
		if err != nil {
			log.Error().Err(err).Msg("error reading packetID")
			return
		}

		if packetID != packet.HandshakeID {
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

	host, addr, err := s.resolveServerAddress(h.ServerAddress)
	if err != nil {
		log.Error().Err(err).Msg("error resolving tcp address")
	}
	log.Info().Str("hostPort", addr.String()).Msg("found backend for connection")

	remote, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Error().Err(err).Msg("unable to connect to backend")
		return
	}

	defer func() {
		if err := remote.Close(); err != nil {
			s.log.Error().Err(err).Str("client", frontendConn.RemoteAddr().String()).Msg("error closing remote connection")
		}
	}()

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
