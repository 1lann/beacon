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
	"time"
)

// Player is the container for the information of a player and their
// connection. See OnConnect.
type Player struct {
	IPAddress      string
	Username       string
	Hostname       string
	ShouldClose    bool
	ForwardAddress string
	InitialPacket  *protocol.Packet
	State          int
	Stream         protocol.Stream
	Connection     net.Conn
}

// A Handler is used for handling when a player attempts to connect to the
// server. See Handle.
type Handler func(player *Player) (message string)

var statuses = make(map[string]*ping.Status)
var handlers = make(map[string]Handler)
var forwarders = make(map[string]string)
var listener net.Listener

// OnForwardConnect is called whenever a connection is forwarded to
// the given IP address, excluding server list pings.
var OnForwardConnect func(ipAddress string)

// OnForwardDisconnect is called when a forwarded connection is closed from
// the given IP address, excluding server list pings.
var OnForwardDisconnect func(ipAddress string, duration time.Duration)

// Stop stops the listener and causes Listen to return.
func Stop() {
	listener.Close()
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

// SetStatus sets the current status that is to be displayed on the
// server list for the given matching hostnames.
func SetStatus(hostnames []string, status *ping.Status) {
	for _, hostname := range hostnames {
		hostname = strings.ToLower(hostname)

		statuses[hostname] = status
	}
}

// ClearStatus clears the current status that was to be displayed on the
// server list for the given matching hostnames.
func ClearStatus(hostnames []string) {
	for _, hostname := range hostnames {
		hostname = strings.ToLower(hostname)

		if _, found := statuses[hostname]; found {
			delete(statuses, hostname)
		}
	}
}

// Handle sets the handler function that is called when a player attempts
// to connect to the server with the given list of hostnames. The function
// should return the message to be displayed to the player. Overrides any
// handlers set by Forward.
func Handle(hostnames []string, handler Handler) {
	for _, hostname := range hostnames {
		hostname = strings.ToLower(hostname)

		if _, found := forwarders[hostname]; found {
			delete(forwarders, hostname)
		}
		handlers[hostname] = handler
	}
}

// Forward forwards the connection to the specified address when a player
// attempts to connect to the server with the given list of hostnames.
// The address MUST include the port number (usually 25565).
// Overrides any handlers set by Handle, and also forwards any server
// list status requests, but does NOT override any statuses stored.
// If you call Handle again, the previously used Status will be used.
func Forward(hostnames []string, address string) {
	for _, hostname := range hostnames {
		hostname = strings.ToLower(hostname)

		if _, found := handlers[hostname]; found {
			delete(handlers, hostname)
		}
		forwarders[hostname] = address
	}
}

// ClearHandlers clears any connection handlers from Handle and Forward
// binded to the given list of hostnames.
func ClearHandlers(hostnames []string) {
	for _, hostname := range hostnames {
		hostname = strings.ToLower(hostname)

		if _, found := handlers[hostname]; found {
			delete(handlers, hostname)
		}

		if _, found := forwarders[hostname]; found {
			delete(forwarders, hostname)
		}
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	player := &Player{
		IPAddress:   strings.Split(conn.RemoteAddr().String(), ":")[0],
		Stream:      protocol.NewStream(conn),
		Connection:  conn,
		ShouldClose: false,
		State:       1,
	}

packetLoop:
	for {
		if player.ShouldClose {
			return
		}

		packetStream, _, err := player.Stream.GetPacketStream()

		if err != nil {
			if err == io.EOF {
				return
			}

			log.Println("beacon: Failed to read next packet:", err)
			return
		}

		packetID, err := packetStream.ReadVarInt()
		if err != nil {
			log.Println("beacon: Failed to read packet ID:", err)
			return
		}

		switch packetID {
		case 0:
			err := handlePacketID0(player, packetStream)
			if err != nil {
				log.Println("beacon: Failed to handle packet ID 0:", err)
			}

			if player.ForwardAddress != "" {
				break packetLoop
			}
		case 1:
			if err := handlePacketID1(player, packetStream); err != nil {
				log.Println("beacon: Failed to handle packet ID 1:", err)
			}
		case 122:
			return
		default:
			log.Println("beacon: Unknown packet ID:", packetID)
		}

		numBytes, err := packetStream.ExhaustPacket()
		if err != nil {
			log.Println("beacon: Failed to exahust", numBytes, "packets:", err)
		} else if numBytes > 0 {
			log.Println("packet id:", packetID)
			log.Println("beacon: Exhausted", numBytes,
				"bytes. (Exhausting packets shouldn't happen).")
		}
	}

	forwardConnection(player)
}

func handlePacketID0(player *Player, ps protocol.PacketStream) error {
	if ps.GetRemainingBytes() == 0 {
		if player.State != 1 {
			return nil
		}

		status, found := statuses[player.Hostname]
		if !found {
			player.ShouldClose = true
			return nil
		}

		err := ping.WriteHandshakeResponse(ps.Stream, *status)
		if err != nil {
			return err
		}

		return nil
	}

	switch player.State {
	case 1:
		handshake, err := ping.ReadHandshakePacket(ps.Stream)
		if err != nil {
			log.Println("beacon: Handshake packet read error:", err)
		}

		player.Hostname = strings.ToLower(handshake.ServerAddress)

		if address, found := forwarders[player.Hostname]; found {
			// Write the handshake data
			initialPacket := protocol.NewPacketWithID(0x00)
			initialPacket.WriteVarInt(handshake.ProtocolVersion)
			initialPacket.WriteString(handshake.ServerAddress)
			initialPacket.WriteUInt16(handshake.ServerPort)
			initialPacket.WriteVarInt(handshake.NextState)
			player.InitialPacket = initialPacket
			player.ForwardAddress = address
			player.State = handshake.NextState
			return nil
		}

		if handshake.NextState == 1 {
			player.State = 1
		} else if handshake.NextState == 2 {
			player.State = 2
		}
	case 2:
		username, err := ps.ReadString()
		if err != nil {
			return err
		}

		player.Username = username

		handler, found := handlers[player.Hostname]
		if !found {
			log.Println("beacon: Missing handler for hostname: " +
				player.Hostname)
			err := ping.DisplayMessage(ps.Stream,
				"Connection rejected. There is no server on this hostname.")
			if err != nil {
				return err
			}

			return nil
		}

		err = ping.DisplayMessage(ps.Stream, handler(player))
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

	status, found := statuses[player.Hostname]
	if !found {
		status = &ping.Status{
			ShowConnection: false,
		}
	}

	return ping.HandlePingPacket(ps.Stream, *status)
}

func forwardConnection(player *Player) {
	remoteConn, err := net.Dial("tcp", player.ForwardAddress)
	if err != nil {
		log.Println("beacon: Failed to connect to remote:", err)
		return
	}

	if OnForwardConnect != nil && player.State == 2 {
		go OnForwardConnect(player.ForwardAddress)
		startTime := time.Now()

		if OnForwardDisconnect != nil {
			defer func() {
				go OnForwardDisconnect(player.ForwardAddress,
					time.Now().Sub(startTime))
			}()
		}
	}

	defer remoteConn.Close()

	lengthPacket := &protocol.Packet{}
	lengthPacket.WriteVarInt(len(player.InitialPacket.Data))

	_, err = remoteConn.Write(append(lengthPacket.Data,
		player.InitialPacket.Data...))
	if err != nil {
		return
	}

	connChannel := make(chan bool)

	go func() {
		io.Copy(remoteConn, player.Connection)
		connChannel <- true
	}()

	go func() {
		io.Copy(player.Connection, remoteConn)
		connChannel <- true
	}()

	<-connChannel
}
