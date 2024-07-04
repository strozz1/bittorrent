package protocol

import (
	"bittorrent/src/types"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
)

type Type int8

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
		return ""
	}
}

type Connection struct {
	con net.Conn
}

/*Creates a TCP connection to the address and returns the Connection struct*/
func CreateConnection(address string) (Connection, error) {
	con, e := net.Dial("tcp", address)
	if e != nil {
		return Connection{}, e
	}
	return Connection{
		con: con,
	}, nil
}
func (c *Connection) Interested() (int, error) {
	content := []byte{}
	content = append(content, []byte{0, 0, 0, 1}...)
	content = append(content, byte(INTERESTED))
	return c.con.Write(content)
}

/* Sends a request message
* @return number of bytes sent or error
 */
func (c *Connection) Request(payload types.PeerRequest) (int, error) {
	msg := peerRequestToBytes(payload)
	return c.con.Write(msg)
}

/* Makes the handshake to the connection with the peer message.
* @returns a tuple with the peer id or the error
 */
func (c *Connection) Handshake(handshake types.PeerHandshake) (string, error) {
	msg := peerHandshakeToBytes(handshake)
	_, e := c.con.Write(msg)
	if e != nil {
		return "", e
	}
	buffer := make([]byte, 68)
	_, e = c.con.Read(buffer)
	if e != nil {
		return "", e
	}

	peerId := buffer[len(buffer)-20:]
	hexadecimalPeerId := fmt.Sprintf("%x", peerId)

	return hexadecimalPeerId, nil
}

func (c *Connection) DownloadPiece(index int, info types.Info) ([]byte,int, error) {


	pieceSize := info.PieceLength
	numPieces := int(math.Ceil(float64(info.Length) / float64(pieceSize)))
	if index == numPieces-1 {
		pieceSize = info.Length - (numPieces-1)*pieceSize //if last piece size is remaining bytes
	}

	blockSize := 16 * 1024 //block size
	numBlocks := int(math.Ceil(float64(pieceSize) / float64(blockSize)))
    piece:=types.NewPiece(uint32(index),uint32(pieceSize))
    for i := range numBlocks {
		currentSize := blockSize
		if i == numBlocks -1 { //if last block
			currentSize = int(pieceSize - blockSize*i) // because of the offset for begin
		}
        //send request
        requestMsg := CreatePeerRequest(uint32(index),uint32(i * blockSize),uint32(currentSize))
		content := peerRequestToBytes(requestMsg)
		c.con.Write(content)
        response, e := c.WaitResponse()
		if e != nil {
			return []byte{},0, e
		}
		pieceResponse := types.BytesToPiece(response)
        piece.AddBlock(pieceResponse.Block,i)



	}
    if !piece.IsComplete() {
        return []byte{},pieceSize,errors.New("The piece is not complete")
    }
    pieceData:=piece.ConcatBlocks()
    return pieceData,pieceSize,nil

}
func (c *Connection) WaitResponse() ([]byte, error) {
    prefixBuffer:=make([]byte,4)
	_, e := c.con.Read(prefixBuffer)
	if e != nil {
		return []byte{}, e
	}
    length:=binary.BigEndian.Uint32(prefixBuffer)


	buffer := make([]byte, length)
	_, e = io.ReadFull(c.con,buffer)
	if e != nil {
		return  []byte{}, e
	}
    

    return buffer, nil

}

type PeerMessage struct {
	lengthPrefix uint32
	id           uint8
	index        uint32
	begin        uint32
	length       uint32
}
