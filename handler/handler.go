// Package handler is a wrapper around the ping package to listen and handle
// connections for displaying information to Minecraft clients.
package handler

import (
	"github.com/1lann/beacon/ping"
	"github.com/1lann/beacon/protocol"
	"io"
	"log"
	"net"
	"strings"
)

// Player is the container for the information of a player and their
// connection. See OnConnect.
type Player struct {
	IP         string
	Username   string
	State      int
	Stream     protocol.Stream
	Connection net.Conn
}

// CurrentStatus is the information the server responds with if a handshake
// is requested, and will be displayed on the client's server list.
var CurrentStatus ping.Status

// OnConnect is the callback function that is called when a player attempts
// to connect to the server. The function should return the message to be
// displayed to the player.
var OnConnect func(player *Player) (message string)

var listener net.Listener

// Stop stops the listener and causes Listen to return.
func Stop() {
	_ = listener.Close()
}

// Listen listens on the specified port to serve Minecraft protocol requests.
func Listen(port string) error {
	var err error
	listener, err = net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(),
				"use of closed network connection") || err == io.EOF {
				return nil
			}

			return err
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	player := &Player{
		IP:         strings.Split(conn.RemoteAddr().String(), ":")[0],
		Stream:     protocol.NewStream(conn),
		Connection: conn,
		State:      1,
	}

	for {
		packetStream, _, err := player.Stream.GetPacketStream()

		if err != nil {
			if err == io.EOF {
				return
			}

			log.Println("Failed to read next packet:", err)
			return
		}

		packetId, err := packetStream.ReadVarInt()
		if err != nil {
			log.Println("Failed to read packet ID:", err)
			return
		}

		switch packetId {
		case 0:
			if err := handlePacketID0(player, packetStream); err != nil {
				log.Println("Failed to handle packet ID 0:", err)
			}
		case 1:
			if err := handlePacketID1(player, packetStream); err != nil {
				log.Println("Failed to handle packet ID 1:", err)
			}
		default:
			log.Println("Unknown packet ID:", packetId)
		}

		numBytes, err := packetStream.ExhaustPacket()
		if err != nil {
			log.Println("Failed to exahust", numBytes, "packets:", err)
		} else if numBytes > 0 {
			log.Println("Exhausted", numBytes, "bytes. (Exhausting packets shouldn't happen).")
		}
	}
}

func handlePacketID0(player *Player, ps protocol.PacketStream) error {
	if ps.GetRemainingBytes() == 0 {
		return nil
	}

	switch player.State {
	case 1:
		handshake, err := ping.ReadHandshakePacket(ps.Stream)
		if err != nil {
			log.Println("Handshake packet read error:", err)
		}

		if handshake.NextState == 1 {
			err := ping.WriteHandshakeResponse(ps.Stream, CurrentStatus)
			if err != nil {
				return err
			}
		} else if handshake.NextState == 2 {
			player.State = 2
		}
	case 2:
		username, err := ps.ReadString()
		if err != nil {
			return err
		}

		player.Username = username

		err = ping.DisplayMessage(ps.Stream, OnConnect(player))
		if err != nil {
			return err
		}
	}

	return nil
}

func handlePacketID1(player *Player, ps protocol.PacketStream) error {
	if ps.GetRemainingBytes() == 0 {
		return nil
	}

	return ping.HandlePingPacket(ps.Stream, CurrentStatus)
}
