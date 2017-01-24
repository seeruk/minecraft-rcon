package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
)

const (
	PACKET_ID_BAD_AUTH     = -1
	PAYLOAD_MAX_SIZE       = 1460
	SERVERDATA_AUTH        = 3
	SERVERDATA_EXECCOMMAND = 2
)

type payload struct {
	packetId   int32  // 4 bytes
	packetType int32  // 4 bytes
	packetBody []byte // Varies
}

func (p *payload) calculatePacketSize() int32 {
	return int32(len(p.packetBody) + 10)
}

func NewClient(host string, port int, pass string) (Client, error) {
	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return Client{}, err
	}

	client := Client{}
	client.connection = conn

	err = client.SendAuthentication(pass)
	if err != nil {
		return Client{}, err
	}

	return client, nil
}

// RCON client based around the Valve RCON Protocol, see more about the protocol in the Valve Wiki
// https://developer.valvesoftware.com/wiki/Source_RCON_Protocol
type Client struct {
	connection net.Conn
}

func (c *Client) SendAuthentication(pass string) error {
	payload := createPayload(SERVERDATA_AUTH, pass)

	_, err := c.sendPayload(payload)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SendCommand(command string) (string, error) {
	payload := createPayload(SERVERDATA_EXECCOMMAND, command)

	response, err := c.sendPayload(payload)
	if err != nil {
		return "", err
	}

	return string(response.packetBody), nil
}

func (c *Client) sendPayload(request payload) (payload, error) {
	packet, err := createPacketFromPayload(request)
	if err != nil {
		return payload{}, err
	}

	_, err = c.connection.Write(packet)
	if err != nil {
		return payload{}, err
	}

	repsonse, err := createPayloadFromPacket(c.connection)
	if err != nil {
		return payload{}, err
	}

	if repsonse.packetId == PACKET_ID_BAD_AUTH {
		return payload{}, errors.New("Authentication unsuccessful")
	}

	return repsonse, nil
}

func createPacketFromPayload(payload payload) ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, payload.calculatePacketSize())
	binary.Write(&buf, binary.LittleEndian, payload.packetId)
	binary.Write(&buf, binary.LittleEndian, payload.packetType)
	binary.Write(&buf, binary.LittleEndian, payload.packetBody)
	binary.Write(&buf, binary.LittleEndian, [2]byte{})

	if buf.Len() >= PAYLOAD_MAX_SIZE {
		return nil, fmt.Errorf("Payload exceeded maximum allowed size of %d.", PAYLOAD_MAX_SIZE)
	}

	return buf.Bytes(), nil
}

func createPayload(packetType int, body string) payload {
	return payload{
		packetId:   rand.Int31(),
		packetType: int32(packetType),
		packetBody: []byte(body),
	}
}

func createPayloadFromPacket(packetReader io.Reader) (payload, error) {
	var packetSize int32
	var packetId int32
	var packetType int32

	// Read packetSize
	err := binary.Read(packetReader, binary.LittleEndian, &packetSize)
	if err != nil {
		return payload{}, err
	}

	// Read packetId
	err = binary.Read(packetReader, binary.LittleEndian, &packetId)
	if err != nil {
		return payload{}, err
	}

	// Read packetType
	err = binary.Read(packetReader, binary.LittleEndian, &packetType)
	if err != nil {
		return payload{}, err
	}

	// Body size length is packet size without the empty string byte at the end
	packetBody := make([]byte, packetSize-8)

	_, err = io.ReadFull(packetReader, packetBody)
	if err != nil {
		return payload{}, err
	}

	result := payload{}
	result.packetId = packetId
	result.packetType = packetType
	result.packetBody = packetBody

	return result, nil
}
