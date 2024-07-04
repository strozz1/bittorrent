package types

import (
	"bittorrent/src/decoder"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
)

type Info struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"` // Solo para archivos de una sola pieza
	PieceLength int    `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"`
}

type MetaInfo struct {
	Announce string `bencode:"announce"`
	Info     Info   `bencode:"info"`
}

func GetMetaInfo(m map[string]any) (MetaInfo, error) {
	infoMap := m["info"].(map[string]any)
	name := infoMap["name"].(string)
	length := infoMap["length"].(int)
	pLen := infoMap["piece length"].(int)
	pieces := infoMap["pieces"].([]byte)
	info := Info{
		Name:        name,
		Length:      length,
		PieceLength: pLen,
		Pieces:      pieces,
	}
	announce := m["announce"].(string)
	metaInfo := MetaInfo{
		Announce: announce,
		Info:     info,
	}
	return metaInfo, nil
}

func MetaInfoFromFile(path string) (MetaInfo, error) {

	content, e := os.ReadFile(path)
	if e != nil {
		fmt.Println("Error reading file. ", e)
		os.Exit(1)
	}
	decoded_bytes, e := decoder.Decode(content)
	if e != nil {
		return MetaInfo{}, e
	}
	dict, _ := decoded_bytes.(map[string]any)

	return GetMetaInfo(dict)

}

type TrackerResp struct {
	Complete   int64  `bencode:"complete"`
	Incomplete int64  `bencode:"incomplete"`
	Interval   int64  `bencode:"interval"`
	Peers      []byte `bencode:"peers"`
}

type IP struct {
	net.IP
	Port int
}

func (ip IP) String() string {
	return fmt.Sprintf("%v:%d", ip.IP, ip.Port)
}
func IPFromStr(ipstr string) (IP, error) {
	var colonIndex int
	for i, l := range ipstr {
		if l == ':' {
			colonIndex = i
			break
		}
	}
	ip := ipstr[:colonIndex]
	port, _ := strconv.Atoi(ipstr[colonIndex+1:])

	IP := IP{
		net.ParseIP(ip),
		port,
	}
	return IP, nil
}

var PREFIX = []byte{0, 0, 0, 1}

type PeerMessage struct {
	Prefix uint32
	Type   int8 `json:"type"`
}

type PeerHandshake struct {
	Prefix   uint32
	Length   int8   `json:"length"`
	Protocol string `json:"protocol"`
	InfoHash string `json:"info_hash"`
	PeerId   string `json:"peer_id"`
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
	Index  uint32
	Length uint32
	Blocks [][]byte
    DownloadedBlocks int
}
func (p *Piece) IsComplete() bool{
    return p.DownloadedBlocks == len(p.Blocks)
}

var blockSize = uint32(16 * 1024) 
func NewPiece(index uint32,length uint32) Piece{
    numBlocks:= (length + blockSize-1) / blockSize
    blocks:=make([][]byte,numBlocks)
    return Piece{
        Index: index,
        Length: length,
        Blocks: blocks,
        DownloadedBlocks: 0,
    }
}
func (p *Piece) AddBlock(block []byte,blockIndex int){
    p.Blocks[blockIndex]=block
    p.DownloadedBlocks++
}

func (p *Piece) ConcatBlocks()[]byte{
    piece:=make([]byte,p.Length)
    offset:=0
    for _,block := range p.Blocks{
        copy(piece[offset:],block)
        offset+=len(block)
    }
    return piece
}
func BytesToPiece(content []byte) PieceResponse {
    content=content[1:]

	return PieceResponse{
		Index: binary.BigEndian.Uint32(content[:4]),
		Begin: binary.BigEndian.Uint32(content[4:8]),
		Block: content[8:],
	}

}
