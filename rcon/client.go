package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/seeruk/minecraft-rcon/errhandling"
)

const (
	PacketIDBadAuth       = -1
	PayloadMaxSize        = 1460
	ServerdataAuth        = 3
	ServerdataExeccommand = 2
)

type payload struct {
	packetID   int32  // 4 bytes
	packetType int32  // 4 bytes
	packetBody []byte // Varies
}

func (p *payload) calculatePacketSize() int32 {
	return int32(len(p.packetBody) + 10)
}

func NewClient(host string, port int, pass string) (*Client, error) {
	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return nil, err
	}

	client := new(Client)
	client.connection = conn
	client.password = pass

	err = client.SendAuthentication(pass)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Client is an RCON client based around the Valve RCON Protocol, see more about the protocol in the
// Valve Wiki: https://developer.valvesoftware.com/wiki/Source_RCON_Protocol
type Client struct {
	connection net.Conn
	password   string
}

func (c *Client) SendAuthentication(pass string) error {
	payload := createPayload(ServerdataAuth, pass)

	_, err := c.sendPayload(payload)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SendCommand(command string) (string, error) {
	payload := createPayload(ServerdataExeccommand, command)

	response, err := c.sendPayload(payload)
	if err != nil {
		return "", err
	}

	// Trim null bytes
	response.packetBody = bytes.Trim(response.packetBody, "\x00")

	return strings.TrimSpace(string(response.packetBody)), nil
}

func (c *Client) Reconnect() error {
	conn, err := net.DialTimeout("tcp",
		c.connection.RemoteAddr().String(), 10*time.Second)
	if err != nil {
		return err
	}

	c.connection = conn

	err = c.SendAuthentication(c.password)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) sendPayload(request *payload) (*payload, error) {
	packet, err := createPacketFromPayload(request)
	if err != nil {
		return nil, err
	}

	_, err = c.connection.Write(packet)
	if err != nil {
		return nil, err
	}

	response, err := createPayloadFromPacket(c.connection)
	if err != nil {
		return nil, err
	}

	if response.packetID == PacketIDBadAuth {
		return nil, errors.New("Authentication unsuccessful")
	}

	return response, nil
}

func createPacketFromPayload(payload *payload) ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, payload.calculatePacketSize())
	binary.Write(&buf, binary.LittleEndian, payload.packetID)
	binary.Write(&buf, binary.LittleEndian, payload.packetType)
	binary.Write(&buf, binary.LittleEndian, payload.packetBody)
	binary.Write(&buf, binary.LittleEndian, [2]byte{})

	if buf.Len() >= PayloadMaxSize {
		return nil, fmt.Errorf("Payload exceeded maximum allowed size of %d.", PayloadMaxSize)
	}

	return buf.Bytes(), nil
}

func createPayload(packetType int, body string) *payload {
	return &payload{
		packetID:   rand.Int31(),
		packetType: int32(packetType),
		packetBody: []byte(body),
	}
}

func createPayloadFromPacket(packetReader io.Reader) (*payload, error) {
	var packetSize int32
	var packetID int32
	var packetType int32

	errs := errhandling.NewStack()

	// Read packetSize, packetID, and packetType
	errs.Add(binary.Read(packetReader, binary.LittleEndian, &packetSize))
	errs.Add(binary.Read(packetReader, binary.LittleEndian, &packetID))
	errs.Add(binary.Read(packetReader, binary.LittleEndian, &packetType))

	if !errs.Empty() {
		return nil, errors.New("createPayloadFromPacket: Failed reading bytes")
	}

	// Body size length is packet size without the empty string byte at the end
	packetBody := make([]byte, packetSize-8)

	_, err := io.ReadFull(packetReader, packetBody)
	if err != nil {
		return nil, err
	}

	result := new(payload)
	result.packetID = packetID
	result.packetType = packetType
	result.packetBody = packetBody

	return result, nil
}
