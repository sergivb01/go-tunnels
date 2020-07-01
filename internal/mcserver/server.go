package mcserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	// only used when debug is enabled in config
	_ "net/http/pprof" // #nosec
	"os"
	"time"

	"github.com/juju/ratelimit"
	"github.com/rs/zerolog"

	"github.com/minebreach/go-tunnels/pkg/mcproto"
	"github.com/minebreach/go-tunnels/pkg/mcproto/packet"
)

const handshakeTimeout = 1 * time.Second

var noDeadline time.Time

// MCServer defines a Minecraft relay server
type MCServer struct {
	log         zerolog.Logger
	packetCoder *mcproto.PacketCodec
	proxies     *roundRobinSwitcher
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

	proxies, err := RoundRobinProxySwitcher(cfg.Proxies)
	if err != nil {
		return nil, fmt.Errorf("loading proxies: %w", err)
	}

	return &MCServer{
		log: zerolog.New(w).With().Str("hostname", hostName).
			Timestamp().Logger().Level(logLevel),
		packetCoder: mcproto.NewPacketCodec(),
		proxies:     proxies,
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

	if s.cfg.Debug {
		go func() {
			s.log.Info().Str("listenAddress", "localhost:6060").Msg("listening for http pprof")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil { // #nosec G108
				s.log.Fatal().Err(err).Msg("listening http for pprof")
			}
		}()
	}

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

func (s *MCServer) handleConnection(ctx context.Context, conn *net.TCPConn, t time.Time) {
	defer func() {
		if err := conn.Close(); err != nil {
			s.log.Warn().Err(err).Str("client", conn.RemoteAddr().String()).Msg("closing client connection")
		}
	}()
	log := s.log.With().Str("client", conn.RemoteAddr().String()).Logger()

	if err := conn.SetNoDelay(true); err != nil {
		log.Warn().Err(err).Msg("setting TCPNoDelay to conn")
	}

	if err := conn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		log.Error().Err(err).Msg("set read deadline")
		return
	}

	packetID, err := s.packetCoder.ReadPacket(conn)
	if err != nil {
		log.Error().Err(err).Msg("reading packetID")
		return
	}

	if packetID != packet.HandshakeID {
		log.Warn().Int("packetID", packetID).Msg("unexpected first packet")
		return
	}

	h := &packet.Handshake{}
	if err := h.Decode(conn); err != nil {
		log.Error().Err(err).Msg("reading handshake")
		return
	}

	if h.State == 1 {
		// TODO: save this packet in const and send it to conn on error
		if err := s.packetCoder.WritePacket(conn, &packet.ServerStatus{
			ServerName: "Minebreach",
			Protocol:   h.ProtocolVersion,
			Motd:       "§3§lMineBreach Tunnels\n§cError - Unknown hostname",
			Favicon:    "",
		}); err != nil {
			log.Error().Err(err).Msg("sending custom ServerListPing response")
		}

		// TODO: ping is not working correctly...
		packetID, err := s.packetCoder.ReadPacket(conn)
		if err != nil {
			log.Error().Err(err).Msg("reading packet after ServerStatus")
			return
		}

		log.Info().Int("packetID", packetID).Int("pingPacketID", packet.PingID).Msg("received packet after server status")

		p := &packet.Ping{}
		if err := p.Decode(conn); err != nil {
			log.Error().Err(err).Msg("reading ping packet")
			return
		}

		if err := s.packetCoder.WritePacket(conn, p); err != nil {
			log.Error().Err(err).Msg("sending ping packet")
		}

		return
	}

	if err = conn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		log.Error().Err(err).Msg("setting deadline before login")
		return
	}

	login := &packet.LoginStart{}
	packetID, err = s.packetCoder.ReadPacket(conn)
	if err != nil {
		log.Error().Err(err).Msg("reading packetID")
		return
	}

	if packetID != packet.HandshakeID {
		log.Warn().Int("packetID", packetID).Msg("unexpected packet after Handshake in LoginState")
		return
	}

	if err := login.Decode(conn); err != nil {
		log.Error().Err(err).Msg("decoding LoginStart")
		return
	}
	log = log.With().Str("playerName", login.Name).Logger()
	log.Debug().Str("playerName", login.Name).Msg("read playerName from LoginStart")

	if err = conn.SetReadDeadline(noDeadline); err != nil {
		log.Error().Err(err).Msg("clear read deadline")
		return
	}

	host, addr, err := s.resolveServerAddress(h.ServerAddress)
	if err != nil {
		log.Error().Err(err).Str("serverAddress", h.ServerAddress).Msg("resolving tcp address")
		return
	}
	log = log.With().Str("host", host).Logger()
	log.Debug().Str("hostPort", addr.String()).Msg("found backend for connection")

	dialer, proxyURL, err := s.proxies.GetProxy()
	if err != nil {
		s.kickError(conn, fmt.Sprintf("Proxy Error %s", proxyURL), err)
		log.Error().Err(err).Msg("getting dialer for proxy")
		return
	}
	log.Debug().Str("proxy", proxyURL).Msg("received dialer for proxy")
	log = log.With().Str("proxy", proxyURL).Logger()

	remote, err := dialer.Dial("tcp", addr.String())
	if err != nil {
		s.kickError(conn, fmt.Sprintf("Proxy Error %s", proxyURL), err)
		log.Error().Err(err).Msg("unable to connect to backend")
		return
	}

	defer func() {
		if err := remote.Close(); err != nil {
			s.log.Warn().Err(err).Str("client", conn.RemoteAddr().String()).Msg("closing remote connection")
		}
	}()

	tcpConn, ok := remote.(*net.TCPConn)
	if !ok {
		log.Error().Err(err).Msg("backend connection is not *net.TCPConn")
		return
	}

	if err := tcpConn.SetNoDelay(true); err != nil {
		log.Warn().Err(err).Msg("setting TCPNoDelay to remote")
	}

	h.ServerAddress = host
	if err := s.packetCoder.WritePacket(tcpConn, h); err != nil {
		log.Error().Err(err).Msg("relay handshake!")
		return
	}

	if err := s.packetCoder.WritePacket(tcpConn, login); err != nil {
		log.Error().Err(err).Msg("relay login!")
		return
	}

	log.Info().Str("took", time.Since(t).String()).Msg("pipe with remote started")
	s.pumpConnections(ctx, conn, tcpConn)
	log.Info().Str("sessionDuration", time.Since(t).String()).Msg("pipe with remote closed")
}

func (s *MCServer) kickError(w io.Writer, reason string, err error) {
	_ = s.packetCoder.WritePacket(w, &packet.LoginDisconnect{
		Reason: "[\"\",{\"text\":\"Minebreach\",\"bold\":true,\"color\":\"blue\"},{\"text\":\" Tunnels\\n\"},{\"text\":\"Error! " + reason + ":\",\"color\":\"red\"},{\"text\":\"\\n\"},{\"text\":\"" + err.Error() + "\",\"color\":\"yellow\"}]",
	})
}
