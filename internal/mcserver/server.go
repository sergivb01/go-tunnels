package mcserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/juju/ratelimit"
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
	cfg         Config
}

// NewConnector creates a new MCServer
func NewConnector(configFile string) (*MCServer, error) {
	cfg, err := readFromFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}

	logLevel := zerolog.InfoLevel
	if cfg.Debug {
		logLevel = zerolog.DebugLevel
	}

	var w io.Writer = &zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	if cfg.Production {
		w = os.Stdout
	}

	hostName, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("getting hostname: %w", err)
	}

	return &MCServer{
		log: zerolog.New(w).With().Str("hostname", hostName).
			Timestamp().Logger().Level(logLevel),
		packetCoder: proto.NewPacketCodec(),
		cfg:         *cfg,
	}, nil
}

// Start starts listening for new connections
func (s *MCServer) Start(ctx context.Context) error {
	addr, err := net.ResolveTCPAddr("tcp", s.cfg.Listen)
	if err != nil {
		return fmt.Errorf("resolving local adderss: %w", err)
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		s.log.Error().Err(err).Msg("unable to start listening")
		return err
	}
	s.log.Info().Str("listenAddress", s.cfg.Listen).Msg("listening for MC client connections")

	return s.acceptConnections(ctx, ln)
}

func (s *MCServer) acceptConnections(ctx context.Context, ln *net.TCPListener) error {
	bucket := ratelimit.NewBucketWithRate(float64(s.cfg.Ratelimit.Rate), int64(s.cfg.Ratelimit.Capacity))

	for {
		select {
		case <-ctx.Done():
			return ln.Close()

		case <-time.After(bucket.Take(1)):
			conn, err := ln.AcceptTCP()
			if err != nil {
				s.log.Error().Err(err).Str("remoteAddr", conn.RemoteAddr().String()).
					Msg("accepting connection")
			} else {
				go s.handleConnection(ctx, conn, time.Now())
			}
		}
	}
}

func (s *MCServer) handleConnection(ctx context.Context, frontendConn *net.TCPConn, t time.Time) {
	defer func() {
		if err := frontendConn.Close(); err != nil {
			s.log.Warn().Err(err).Str("client", frontendConn.RemoteAddr().String()).Msg("closing frontend connection")
		}
	}()
	log := s.log.With().Str("client", frontendConn.RemoteAddr().String()).Logger()

	if err := frontendConn.SetNoDelay(true); err != nil {
		log.Warn().Err(err).Msg("setting TCPNoDelay to frontendConn")
	}

	if err := frontendConn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		log.Error().Err(err).Msg("failed to set read deadline")
		return
	}

	packetID, err := s.packetCoder.ReadPacket(frontendConn)
	if err != nil {
		log.Error().Err(err).Msg("reading packetID")
		return
	}

	if packetID != packet.HandshakeID {
		log.Warn().Int("packetID", packetID).Msg("unexpected first packet")
		return
	}

	h := &packet.Handshake{}
	if err := h.Decode(frontendConn); err != nil {
		log.Error().Err(err).Msg("reading handshake")
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
			log.Error().Err(err).Msg("sending custom ServerListPing response")
		}

		// TODO: check for packetID == packet.PingID
		_, err := s.packetCoder.ReadPacket(frontendConn)
		if err != nil {
			log.Error().Err(err).Msg("reading packet after ServerStatus")
			return
		}

		p := &packet.Ping{}
		if err := p.Decode(frontendConn); err != nil {
			log.Error().Err(err).Msg("reading ping packet")
			return
		}

		if err := s.packetCoder.WritePacket(frontendConn, p); err != nil {
			log.Error().Err(err).Msg("sending ping packet")
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
	log := s.log.With().Str("client", frontendConn.RemoteAddr().String()).Logger()

	login := &packet.LoginStart{}
	// TODO: cleanup logic
	if h.State == 2 {
		packetID, err := s.packetCoder.ReadPacket(frontendConn)
		if err != nil {
			log.Error().Err(err).Msg("reading packetID")
			return
		}

		if packetID != packet.HandshakeID {
			log.Warn().Int("packetID", packetID).Msg("unexpected packet after Handshake in LoginState")
			return
		}

		if err := login.Decode(frontendConn); err != nil {
			log.Error().Err(err).Msg("decoding LoginStart")
			return
		}

		log = log.With().Str("playerName", login.Name).Logger()
		log.Debug().Str("playerName", login.Name).Msg("read playerName from LoginStart")
	}

	host, addr, err := s.resolveServerAddress(h.ServerAddress)
	if err != nil {
		log.Error().Err(err).Str("serverAddress", h.ServerAddress).Msg("resolving tcp address")
		return
	}
	log = log.With().Str("host", host).Logger()
	log.Debug().Str("hostPort", addr.String()).Msg("found backend for connection")

	remote, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Error().Err(err).Msg("unable to connect to backend")
		return
	}

	defer func() {
		if err := remote.Close(); err != nil {
			s.log.Warn().Err(err).Str("client", frontendConn.RemoteAddr().String()).Msg("closing remote connection")
		}
	}()

	if err := remote.SetNoDelay(true); err != nil {
		log.Warn().Err(err).Msg("setting TCPNoDelay to remote")
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
	log.Info().Str("sessionDuration", time.Since(t).String()).Msg("pipe with remote closed")
}

// func (s *MCServer) kickError(w io.Writer, reason string, err error) {
// 	_ = s.packetCoder.WritePacket(w, &packet.LoginDisconnect{
// 		Reason: "[\"\",{\"text\":\"Minebreach\",\"bold\":true,\"color\":\"blue\"},{\"text\":\" Tunnels\\n\"},{\"text\":\"Error! " + reason + ":\",\"color\":\"red\"},{\"text\":\"\\n\"},{\"text\":\"" + err.Error() + "\",\"color\":\"yellow\"}]",
// 	})
// }
