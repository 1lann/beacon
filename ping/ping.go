// Package ping is used for reading and sending data regarding
// handshaking, pinging and connecting.
// It follows the Minecraft protocol version 5.
package ping

import (
	"encoding/json"
	"github.com/1lann/beacon/protocol"
)

// Internally used struct for JSON serialization.
type StatusResponse struct {
	Version     `json:"version"`
	Players     `json:"players"`
	Description string `json:"description"`
}

// Internally used struct for JSON serialization.
type Version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

// Internally used struct for JSON serialization.
type Players struct {
	Max    int `json:"max"`
	Online int `json:"online"`
}

// HandshakePacket contains the decoded data from a handshake packet.
// See ReadHandshakePacket.
type HandshakePacket struct {
	ProtocolVersion int
	ServerAddress   string
	ServerPort      uint16
	NextState       int
}

// Status is the container for the information to respond with
// on the Minecraft server list menu.
type Status struct {
	OnlinePlayers  int
	MaxPlayers     int
	Message        string
	ShowConnection bool
}

// ReadHandshakePacket reads a handshake packet (packet ID 0) and decodes it.
func ReadHandshakePacket(s protocol.Stream) (HandshakePacket, error) {
	handshake := HandshakePacket{}
	var err error
	if handshake.ProtocolVersion, err = s.ReadVarInt(); err != nil {
		return HandshakePacket{}, err
	}

	if handshake.ServerAddress, err = s.ReadString(); err != nil {
		return HandshakePacket{}, err
	}

	if handshake.ServerPort, err = s.ReadUInt16(); err != nil {
		return HandshakePacket{}, err
	}

	handshake.NextState, err = s.ReadVarInt()
	return handshake, err
}

// WriteHandshakeResponse writes a response with a status that will be
// displayed on the requesting player's server list menu.
func WriteHandshakeResponse(s protocol.Stream, status Status) error {
	statusResponse := StatusResponse{
		Version: Version{
			Name:     "1.7.10",
			Protocol: 5,
		},
		Players: Players{
			Max:    status.MaxPlayers,
			Online: status.OnlinePlayers,
		},
		Description: status.Message,
	}

	data, err := json.Marshal(statusResponse)
	if err != nil {
		return err
	}

	responsePacket := protocol.NewPacketWithId(0x00)
	responsePacket.WriteString(string(data))
	err = s.WritePacket(responsePacket)
	return err
}

// HandlePingPacket handles a ping packet used by the Minecraft client
// used to measure the round trip time of the connection.
func HandlePingPacket(s protocol.Stream, status Status) error {
	if !status.ShowConnection {
		return nil
	}

	time, err := s.ReadInt64()
	if err != nil {
		return err
	}
	responsePacket := protocol.NewPacketWithId(0x01)
	responsePacket.WriteInt64(time)
	err = s.WritePacket(responsePacket)
	return err
}

// DisplayMessage responds with a disconnect message to the player
// when they attempt to connect to the server.
func DisplayMessage(s protocol.Stream, message string) error {
	responsePacket := protocol.NewPacketWithId(0x00)

	chatMessage := message

	data, err := json.Marshal(chatMessage)
	if err != nil {
		return err
	}

	responsePacket.WriteString(string(data))
	err = s.WritePacket(responsePacket)
	return err
}
