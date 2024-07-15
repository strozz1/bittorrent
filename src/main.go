package main

import (
	"bittorrent/src/decoder"
	"bittorrent/src/protocol"
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {

	command := os.Args[1]
	arg2 := os.Args[2]

	switch command {
	case "decode":
		cmd_decode([]byte(arg2))
	case "info":
		cmd_info(arg2)
	case "peer":
		cmd_peer(arg2)
	case "handshake":
		arg3 := os.Args[3]
		cmd_handshake(arg2, arg3)
	case "download_piece":
		arg3 := os.Args[3]
		arg4, _ := strconv.Atoi(os.Args[4])
		cmdDownloadPiece(arg2, arg3, arg4)
	case "download":
		arg3 := os.Args[3]
		cmdDownload(arg2, arg3)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

}

func cmd_decode(bencode []byte) {
	res, e := decoder.Decode(bencode)
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	Marshall(res)

}
func cmd_info(path string) {
	metaInfo,_, e := decoder.MetaInfoFromFile(path)
    if e!=nil{
        log.Panicln(e)
    }
	print_info(metaInfo)
}

func cmd_peer(path string) {
	metaInfo,_, _ := decoder.MetaInfoFromFile(path)
	p, e := protocol.GetPeers(metaInfo)
	if e != nil {
		fmt.Println("Error", e)
		os.Exit(1)
	}
	for _, peer := range p {
		println(peer.String())
	}
}
func cmd_handshake(path string, ip_str string) {
	_,hash, _ := decoder.MetaInfoFromFile(path)
	handshake := protocol.PeerHandshake{
		Protocol: "BitTorrent protocol",
		InfoHash: string(hash),
		PeerId:   "00112233445566778899",
	}

	con, error := protocol.CreateConnection(ip_str)
	if error != nil {
		fmt.Printf("Error creating connection: %v\n", error)
		return
	}
	peerId, e := con.Handshake(handshake)
	if e != nil {
		fmt.Println("Error: ", e)
	}
	fmt.Println("Peer id: ", peerId)

}
func cmdDownloadPiece(path string, file string, index int) {

	//obtain metainfo and hash
	metaInfo,hash, _ := decoder.MetaInfoFromFile(file)
	//get peers
	peers, _ := protocol.GetPeers(metaInfo)
	if len(peers) == 0 {
		println("No peers found")
		os.Exit(1)
	}

	//create connection
	con, error := protocol.CreateConnection(peers[0].String())
	log.Printf("Connecting to %s\n", peers[0].String())
	if error != nil {
		fmt.Printf("Error creating connection: %v\n", error)
		return
	}
	handshake := protocol.PeerHandshake{
		Protocol: "BitTorrent protocol",
		InfoHash: string(hash),
		PeerId:   "00112233445566778899",
	}
	peerId, _ := con.Handshake(handshake)
	log.Printf("Handshake made, Peer id: %s\n", peerId)

	piece, _, e := con.DownloadPiece(index, metaInfo.Info)

	if e != nil {
		fmt.Printf("Error downloading piece: %v\n", e)
		return
	}
	newFile, _ := os.Create(path)
	defer newFile.Close()

	writer := bufio.NewWriter(newFile)

	_, err := writer.Write(piece)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("Piece", index, "downloaded to", path)

}

func cmdDownload(path string, file string) {

	//obtain metainfo and hash
	metaInfo, hash, _ := decoder.MetaInfoFromFile(file)
	//get peers
	peers, _ := protocol.GetPeers(metaInfo)

	//create connection
	// TODO: check if peers work if more than 1
	con, e := protocol.CreateConnection(peers[0].String())
	log.Println("Connecting to", peers[0].String())
	if e != nil {
		log.Panicln(e)
	}
	handshake := protocol.NewHandshake(hash)
	peerId, _ := con.Handshake(handshake)
	log.Printf("Handshake made, peer_id: %s.\n", peerId)

	piecelist := pieces_hashes(metaInfo.Info.Pieces)

	con.WaitResponse() // unchoke received or bitfield
	_, e = con.SendInterested()         //send interested msg
	if e != nil {
		log.Panicln(e)
	}
	totalDownloaded := 0
	piecesLeft := len(piecelist)

	for piecesLeft > 0 {
        fmt.Printf("\rDownloading pieces: %d/%d...",len(piecelist)-piecesLeft,len(piecelist))

		msgType, content, _ := con.WaitResponse() //wait for Have to tell index
		if msgType == protocol.HAVE {
			indexAvailable := con.ManageResponse(msgType, content).(uint32)
            //todo get piece data
			_, _, e := con.DownloadPiece(int(indexAvailable), metaInfo.Info)
			if e != nil {
				log.Panicln(e)
			}
            piecesLeft--
			//log.Println("Piece", indexAvailable, "downloaded. Pieces remaining", piecesLeft)
			_, e = con.SendHave(indexAvailable) //send we have received the piece
		}else{
            //log.Println("NO have, type",msgType.String())
        }

	}
    log.Println("All pieces received")

	return

	for i, p := range piecelist {
		newFile, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		piece, size, e := con.DownloadPiece(i, metaInfo.Info)
		totalDownloaded += size
		if e != nil {
			fmt.Printf("Error downloading piece: %v\n", e)
			return
		}

		fmt.Printf("Piece %d downloaded: %d bytes. Remaining bytes: %d\n", i, len(piece), metaInfo.Info.Length-totalDownloaded)
		sha := sha1.Sum(piece)
		shaD := hex.EncodeToString(sha[:])
		if p != shaD {
			println("Error: downloaded piece", i, "hash is not the same as actual")
			fmt.Printf("Expected:\n%s\nActual:\n%s\n", p, shaD)
			return
		}

		newFile.Write(piece)
		newFile.Close()
	}

	//println("File", metaInfo.Info.Name, "downloaded to", path)

}

func Marshall(content any) {
	jsonOutput, _ := json.Marshal(content)
	fmt.Println(string(jsonOutput))
}

func pieces_hashes(pieces []byte) []string {
	list := []string{}
	hashLength := 20
	for i := 0; i < len(pieces)/hashLength; i++ {
		list = append(list, hex.EncodeToString([]byte(pieces[i*hashLength:(i+1)*hashLength])))
	}
	return list

}
func print_info(metaInfo decoder.MetaInfo) {
	hash, _ := decoder.CalculateInfoHash(metaInfo.Info)
	fmt.Printf("Tracker URL: %s\n", metaInfo.Announce)
	fmt.Printf("Length: %d\n", metaInfo.Info.Length)
	fmt.Printf("Info Hash: %s\n", hash)
	fmt.Printf("Piece Length: %d\n", metaInfo.Info.PieceLength)
	fmt.Printf("Piece Hashes:\n")
	list := pieces_hashes(metaInfo.Info.Pieces)
	for _, l := range list {
		println(l)
	}

}
