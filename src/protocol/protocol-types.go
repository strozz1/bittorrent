package protocol

import (
	"bytes"
	"encoding/binary"
)

type Type uint8

const (
	CHOKE        Type = 0
	UNCHOKE      Type = 1
	INTERESTED   Type = 2
	NOT_INTEREST Type = 3
	HAVE         Type = 4
	BITFIELD     Type = 5
	REQUEST      Type = 6
	PIECE        Type = 7
	CANCEL       Type = 8
)

type Message interface {
	toBytes() []byte
}



func (t Type) String() string {
	switch t {
	case CHOKE:
		return "CHOKE"
	case UNCHOKE:
		return "UNCHOKE"
	case INTERESTED:
		return "INTERESTED"
	case NOT_INTEREST:
		return "NOT_INTERESTED"
	case HAVE:
		return "HAVE"
	case BITFIELD:
		return "BITFIELD"

	case REQUEST:
		return "REQUEST"
	case PIECE:
		return "PIECE"
	case CANCEL:
		return "CANCEL"
	default:
		return "BITTORRENT"
	}
}

type PeerMessage struct {
	lengthPrefix uint32
	id           uint8
	index        uint32
	begin        uint32
	length       uint32
}

type HaveMessage struct {
	LengthPrefix uint32
	MsgType      Type
	Index        uint32
}

func (msg *HaveMessage) toBytes() []byte {
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, msg)
	return buffer.Bytes()
}

var PREFIX = []byte{0, 0, 0, 1}

type PeerHandshake struct {
	Prefix   uint32
	Length   int8   `json:"length"`
	Protocol string `json:"protocol"`
	InfoHash string `json:"info_hash"`
	PeerId   string `json:"peer_id"`
}
func NewHandshake(hash []byte)PeerHandshake{
   	return PeerHandshake{
		Protocol: "BitTorrent protocol",
		InfoHash: string(hash),
		PeerId:   "00112233445566778899",
	}
}

type PeerRequest struct {
	Prefix uint32
	Type   uint8
	Index  uint32
	Begin  uint32
	Length uint32
}

type PieceResponse struct {
	Index uint32
	Begin uint32
	Block []byte
}
type Piece struct {
	Index            uint32
	Length           uint32
	Blocks           [][]byte
	DownloadedBlocks int
}

func (p *Piece) IsComplete() bool {
	return p.DownloadedBlocks == len(p.Blocks)
}

func NewPiece(index uint32, length uint32) Piece {
	var blockSize = uint32(16 * 1024)
	numBlocks := (length + blockSize - 1) / blockSize
	blocks := make([][]byte, numBlocks)
	return Piece{
		Index:            index,
		Length:           length,
		Blocks:           blocks,
		DownloadedBlocks: 0,
	}
}
func (p *Piece) AddBlock(block []byte, blockIndex int) {
	p.Blocks[blockIndex] = block
	p.DownloadedBlocks++
}

func (p *Piece) ConcatBlocks() []byte {
	piece := make([]byte, p.Length)
	offset := 0
	for _, block := range p.Blocks {
		copy(piece[offset:], block)
		offset += len(block)
	}
	return piece
}
func BytesToPiece(content []byte) PieceResponse {

	return PieceResponse{
		Index: binary.BigEndian.Uint32(content[:4]),
		Begin: binary.BigEndian.Uint32(content[4:8]),
		Block: content[8:],
	}

}
