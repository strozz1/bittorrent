package protocol

import (
	"bittorrent/src/decoder"
	"bittorrent/src/types"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"

	"github.com/jackpal/bencode-go"
)

func GetPeers(metaInfo types.MetaInfo) ([]types.IP, error) {
	hash, _ := CalculateInfoHash(metaInfo.Info)
	hashD, _ := hex.DecodeString(hash)
	params := url.Values{}
	params.Add("info_hash", string(hashD))
	params.Add("peer_id", "00112233445566778899")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", fmt.Sprint(metaInfo.Info.Length))
	params.Add("compact", "1")
	url := fmt.Sprintf("%s?%s", metaInfo.Announce, params.Encode())

	resp, e := http.Get(url)
	if e != nil {
		return nil, e

	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}


	decoded, _ := decoder.Decode(content)


	dict := decoded.(map[string]any)


	interval := dict["interval"].(int)
	comp := dict["complete"].(int)
	incomp := dict["incomplete"].(int)

	peers := dict["peers"].(string)

	tracker := types.TrackerResp{
		Interval:   int64(interval),
		Complete:   int64(comp),
		Incomplete: int64(incomp),
		Peers:      []byte(peers),
	}
	ips, _ := parsePeers(tracker.Peers)
	return ips, nil

}

func parsePeers(peers []byte) ([]types.IP, error) {

	SIZE_IP := 6
	size := len(peers) / SIZE_IP
	ips := make([]types.IP, size)
	for i := 0; i < size; i++ {
		end := SIZE_IP*i + 6
		start := SIZE_IP * i
		ips[i] = types.IP{
			IP:   net.IPv4(peers[start:end][0], peers[start:end][1], peers[start:end][2], peers[start:end][3]),
			Port: int(big.NewInt(0).SetBytes(peers[start:end][4:]).Uint64()),
		}
	}

	return ips, nil

}

// Función para calcular el info hash
func CalculateInfoHash(info types.Info) (string, error) {
	var buffer bytes.Buffer

	err := bencode.Marshal(&buffer, info)
	if err != nil {
		return "", err
	}

	hash := sha1.Sum(buffer.Bytes())

	// Convertir el hash a una cadena hexadecimal
	return hex.EncodeToString(hash[:]), nil
}

func ConnectWithPeer(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
}

/* Transform the PeerHandshake struct to []byte with the specifications from the protocol */
func peerHandshakeToBytes(handshake types.PeerHandshake) []byte {

	content := []byte{}
	content = append(content, 19) // prot size
	content = append(content, handshake.Protocol...)
	for range 8 {
		content = append(content, 0)
	}
	content = append(content, handshake.InfoHash...)
	content = append(content, handshake.PeerId...)

	return content
}

func peerRequestToBytes(payload types.PeerRequest) []byte {
    var buffer bytes.Buffer
    binary.Write(&buffer,binary.BigEndian,payload)

	return buffer.Bytes()

}

func peerMessageToByte(msg types.PeerMessage) []byte {
	var buffer bytes.Buffer
    binary.Write(&buffer,binary.BigEndian,msg)

	return buffer.Bytes()
}

func CreatePeerRequest(index uint32, begin uint32, length uint32) types.PeerRequest{
    return types.PeerRequest{
            Prefix: uint32(13),
            Type: uint8(REQUEST),
			Index:  index,
			Begin:  begin,
			Length: length,
    }

}
