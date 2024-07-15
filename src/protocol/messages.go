package protocol

import (
	"bittorrent/src/decoder"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
)



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
func (c *Connection) SendInterested() (int, error) {
	content := []byte{}
	content = append(content, []byte{0, 0, 0, 1}...)
	content = append(content, byte(INTERESTED))
	return c.con.Write(content)
}

/* Sends a request message
* @return number of bytes sent or error
 */
func (c *Connection) SendRequest(payload PeerRequest) (int, error) {
	msg := peerRequestToBytes(payload)
	//log.Println("PROTOCOL: OUT-> Request")
	return c.con.Write(msg)
}

func (c *Connection) SendHave(index uint32) (int, error) {
	haveMsg := HaveMessage{
		LengthPrefix: 5,
		MsgType:      HAVE,
		Index:        index,
	}
	//log.Println("PROTOCOL: OUT-> HAVE")
	return c.con.Write(haveMsg.toBytes())

}

/* Makes the handshake to the connection with the peer message.
* @returns a tuple with the peer id or the error
 */
func (c *Connection)Handshake(handshake PeerHandshake) (string, error) {
	msg := peerHandshakeToBytes(handshake)
	_, e := c.con.Write(msg)
	//log.Println("PROTOCOL: OUT-> Handshake")
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
	//log.Println("PROTOCOL: IN-> Handshake")

	return hexadecimalPeerId, nil
}

func (c *Connection)DownloadPiece(index int, info decoder.Info) ([]byte, int, error) {

	pieceSize := info.PieceLength
	numPieces := int(math.Ceil(float64(info.Length) / float64(pieceSize)))
	if index == numPieces-1 {
		pieceSize = info.Length - (numPieces-1)*pieceSize //if last piece size is remaining bytes
	}

	blockSize := 16 * 1024 //block size
	numBlocks := int(math.Ceil(float64(pieceSize) / float64(blockSize)))
	piece := NewPiece(uint32(index), uint32(pieceSize))
	//log.Println("Downloading piece", index)
    	for i := range numBlocks {
        fmt.Print("+")

		currentSize := blockSize
		if i == numBlocks-1 { //if last block
			currentSize = int(pieceSize - blockSize*i) // because of the offset for begin
		}
		//send request
		requestMsg := CreatePeerRequest(uint32(index), uint32(i*blockSize), uint32(currentSize))
		content := peerRequestToBytes(requestMsg)
		c.con.Write(content)
		//log.Printf("Request sent to peer: %+v\n", content)
		msgType, response, e := c.WaitResponse()
		if e != nil {
			return []byte{}, 0, e
		}
		var pieceResponse PieceResponse
		if msgType == PIECE {
			pieceResponse = c.ManageResponse(msgType, response).(PieceResponse)
		} else {
			return []byte{}, 0, errors.New("Got another pacakge, not PIECE")
		}
		piece.AddBlock(pieceResponse.Block, i)
		//log.Printf("Block %d downloaded.\n", i)
	}
        fmt.Print("\r\033[K")
    if !piece.IsComplete() {
		return []byte{}, pieceSize, errors.New("The piece is not complete")
	}
	pieceData := piece.ConcatBlocks()
	return pieceData, pieceSize, nil

}
func (c *Connection) WaitResponse() (Type, []byte, error) {
	//log.Println("Waiting for peer response")
	prefixBuffer := make([]byte, 4)
	_, e := c.con.Read(prefixBuffer)
	if e != nil {
		return 0, []byte{}, e
	}
	length := binary.BigEndian.Uint32(prefixBuffer)

    if length == 0 {
        return 9,nil,nil
    }
	buffer := make([]byte, length)
	_, e = io.ReadFull(c.con, buffer)
	if e != nil {
		return 0, []byte{}, e
	}
	msgType := Type(buffer[0])

    //log.Printf("PROTOCOL: IN->%s\n", msgType.String())

	return msgType, buffer[1:], nil

}

func (c *Connection) ManageResponse(msgType Type, buffer []byte) any {
	switch msgType {
	case HAVE:
		index := binary.BigEndian.Uint32(buffer)
		return index
	case PIECE:
		return BytesToPiece(buffer)
	default: //if default return buffer as it is
		return buffer
	}
}


