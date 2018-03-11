package query

import (
	"strconv"
	"strings"

	"github.com/irmine/binutils"
	"net"
)

const (
	Challenge  = 0x09
	Statistics = 0x00
)

// Header is the header of each query.
var Header = []byte{0xfe, 0xfd}

// Query is used to encode/decode queries.
type Query struct {
	*binutils.Stream
	Address string
	Port    uint16

	Header  byte
	QueryId int32
	Token   []byte

	Statistics []byte

	IsShort bool
	Data    []byte
}

// NewQueryFromRaw returns a query from a raw packet.
func NewFromRaw(buffer []byte, addr *net.UDPAddr) *Query {
	var stream = binutils.NewStream()
	stream.Buffer = buffer
	return &Query{stream, addr.IP.String(), uint16(addr.Port), 0, 0, []byte{}, []byte{}, false, []byte{}}
}

// NewQuery returns a new query with an address and port.
func New(address string, port uint16) *Query {
	return &Query{binutils.NewStream(), address, port, 0, 0, []byte{}, []byte{}, false, []byte{}}
}

// DecodeServer decodes the query sent by the client.
func (query *Query) DecodeServer() {
	query.Offset = 2
	query.Header = query.GetByte()
	query.QueryId = query.GetInt()

	if query.Header == Statistics {
		query.Token = query.Get(4)
		var length = len(query.Get(-1)) + 4 // Token size + padding
		if length != 8 {
			query.IsShort = true
		}
	}
}

// EncodeServer encodes the query to send to the client.
func (query *Query) EncodeServer() {
	query.PutByte(query.Header)
	query.PutInt(query.QueryId)

	switch query.Header {
	case Challenge:
		var token = query.Token
		var offset = 0
		var tokenString = strconv.Itoa(int(binutils.ReadInt(&token, &offset)))

		var padding = 12 - len(tokenString)

		query.PutBytes([]byte(tokenString))
		for i := 0; i < padding; i++ {
			query.PutByte(0)
		}
	case Statistics:
		query.PutBytes(query.Statistics)
		query.PutByte(0)
	}
}

// EncodeClient encodes a query to send to the server.
func (query *Query) EncodeClient() {
	query.PutBytes(Header)
	query.PutByte(query.Header)
	query.PutInt(query.QueryId)

	if query.Header == Statistics {
		query.PutBytes(query.Token)
		query.PutBytes([]byte{0, 0, 0, 0})
	}
}

// DecodeClient decodes a query sent by the server.
func (query *Query) DecodeClient() {
	query.Header = query.GetByte()
	query.QueryId = query.GetInt()

	switch query.Header {
	case Challenge:
		var buf []byte
		var i, _ = strconv.ParseInt(strings.TrimRight(string(query.Get(-1)), "\x00"), 0, 32)

		binutils.WriteInt(&buf, int32(i))
		query.Token = buf
	case Statistics:
		query.Data = query.Get(-1)
	}
}
