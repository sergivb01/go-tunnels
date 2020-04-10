package mcserver

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/juju/ratelimit"
	"github.com/rs/zerolog"

	"github.com/sergivb01/mctunnel/internal/protocol"
	"github.com/sergivb01/mctunnel/internal/protocol/chat"
	"github.com/sergivb01/mctunnel/internal/protocol/packet"
	"github.com/sergivb01/mctunnel/internal/protocol/types"
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

	c := protocol.NewConnection(frontendConn)

	host := "abc"
	var port int

	var lastPk packet.Handshake

	for host == "abc" {
		pk, err := c.Next()
		if err == io.EOF || pk == nil {
			return
		}

		rawPacket, err := c.D.Decode(pk)
		if err == protocol.ErrUknownPacket {
			continue
		}

		cLog.Info().Int("packetID", pk.ID).Msg("received packet!")

		switch p := rawPacket.(type) {
		case packet.StatusRequest:
			if _, err := c.Write(getRandomResponse()); err != nil {
				cLog.Error().Err(err).Msg("error trying to write StatusResponse")
				return
			}
			break
		case packet.Handshake:
			c.SetState(protocol.State(p.NextState))
			cLog.Info().Int("state", int(p.NextState)).Msg("STATE IS BLABLABLA")

			if p.NextState == 1 {
				if _, err := c.Write(getRandomResponse()); err != nil {
					cLog.Error().Err(err).Msg("error trying to write StatusResponse")
				}
				return
			}

			host, port, err = ExtractHostPort(string(p.ServerAddress))
			if err != nil {
				cLog.Error().Err(err).Str("server", string(p.ServerAddress)).Msg("could not find backend")
				return
			}
			p.ServerAddress = types.String(host)
			lastPk = p
		// case packet.StatusPing:
		// 	pong := packet.StatusPong{Payload: p.Payload}
		// 	if _, err := c.Write(pong); err != nil {
		// 		cLog.Error().Err(err).Msg("error trying to write StatusPong")
		// 		return
		// 	}
		default:
			cLog.Error().Msg("what the fuck...?")
		}
	}

	s.findAndConnectBackend(ctx, frontendConn, lastPk, host, port)
}

func (s *MCServer) findAndConnectBackend(ctx context.Context, frontendConn net.Conn, lastPk packet.Handshake, host string, port int) {
	// backendHostPort, resolvedHost := Routes.FindBackendForServerAddress(serverAddress)
	// host, port, err := ExtractHostPort(serverAddress)
	backendHostPort := fmt.Sprintf("%s:%d", host, port)
	cLog := s.log.With().Str("client", frontendConn.RemoteAddr().String()).Str("backendHostPort", backendHostPort).Logger()
	cLog.Info().Msg("connecting to backend")

	backendConn, err := net.Dial("tcp", backendHostPort)
	if err != nil {
		cLog.Error().Err(err).Msg("unable to connect to backend")
		return
	}

	time.Sleep(time.Millisecond * 300)

	amount, err := protocol.NewConnection(backendConn).Write(lastPk)
	if err != nil {
		cLog.Error().Err(err).Msg("failed to write handshake to backend")
		return
	}
	cLog.Debug().Int("amout", amount).Msg("relayed handshake to backend")

	if err = frontendConn.SetReadDeadline(noDeadline); err != nil {
		cLog.Error().Err(err).Msg("failed to clear read deadline")
		return
	}

	s.pumpConnections(ctx, frontendConn, backendConn)
}

func getRandomResponse() packet.StatusResponse {
	resp := packet.StatusResponse{}
	resp.Status.Version.Name = "1.8.8"
	resp.Status.Version.Protocol = 47
	resp.Status.Players.Max = rand.Intn(100)
	resp.Status.Players.Online = rand.Intn(101)
	resp.Status.Description = chat.TextComponent{
		Text: "Spoofed!",
		Component: chat.Component{
			Bold:  true,
			Color: chat.ColorRed,
		},
	}
	return resp
}
