package main

import (
	"bittorrent/src/decoder"
	"bittorrent/src/protocol"
	"bittorrent/src/types"
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	metaInfo, _ := types.MetaInfoFromFile(path)
	print_info(metaInfo)
}

func cmd_peer(path string) {
	metaInfo, _ := types.MetaInfoFromFile(path)
    p,e:=protocol.GetPeers(metaInfo)
	if e != nil {
		fmt.Println("Error",e)
		os.Exit(1)
	}
    for _,peer := range p{
        println(peer.String())
    }
}
func cmd_handshake(path string, ip_str string) {
	metaInfo, _ := types.MetaInfoFromFile(path)
	hash, _ := protocol.CalculateInfoHash(metaInfo.Info)
	hash_decoded, _ := hex.DecodeString(hash)

	handshake := types.PeerHandshake{
		Protocol: "BitTorrent protocol",
		InfoHash: string(hash_decoded),
		PeerId:   "00112233445566778899",
	}

	con, error := protocol.CreateConnection(ip_str)
	if error != nil {
		println("Error creating connection: ", error)
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
	metaInfo, _ := types.MetaInfoFromFile(file)
	hash, _ := protocol.CalculateInfoHash(metaInfo.Info)
	hash_decoded, _ := hex.DecodeString(hash)

	//get peers
	peers, _ := protocol.GetPeers(metaInfo)
    if len(peers) == 0{
        println("No peers found")
        os.Exit(1)
    }

	//create connection
	con, error := protocol.CreateConnection(peers[0].String())
	println("Connecting to ", peers[0].String())
	if error != nil {
		println("Error creating connection: ", error)
		return
	}
	handshake := types.PeerHandshake{
		Protocol: "BitTorrent protocol",
		InfoHash: string(hash_decoded),
		PeerId:   "00112233445566778899",
	}
	peerId, _ := con.Handshake(handshake)
	fmt.Printf("Handshake made, Peer id: %s\n", peerId)

	piece,_, e := con.DownloadPiece(index, metaInfo.Info)

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
	println("Piece", index, "downloaded to", path)

}

func cmdDownload(path string, file string) {

	//obtain metainfo and hash
	metaInfo, _ := types.MetaInfoFromFile(file)
	hash, _ := protocol.CalculateInfoHash(metaInfo.Info)
	hash_decoded, _ := hex.DecodeString(hash)

	//get peers
	peers, _ := protocol.GetPeers(metaInfo)

	//create connection
	con, error := protocol.CreateConnection(peers[0].String())
	println("Connecting to ", peers[0].String())
	if error != nil {
		println("Error creating connection: ", error)
		return
	}
	handshake := types.PeerHandshake{
		Protocol: "BitTorrent protocol",
		InfoHash: string(hash_decoded),
		PeerId:   "00112233445566778899",
	}
	peerId, _ := con.Handshake(handshake)
	fmt.Printf("Handshake made, Peer id: %s\n", peerId)

	piecelist := pieces_hashes(metaInfo.Info.Pieces)

    totalDownloaded:=0


    
	
	con.WaitResponse() //wait for bitfield
    _, e := con.Interested() //send interested msg
	if e != nil {
        println("Error: ",e)
	}
    con.WaitResponse()
    
 

	for i, p := range piecelist {
        newFile, _ := os.OpenFile(path,os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		piece,size, e := con.DownloadPiece(i, metaInfo.Info)
        totalDownloaded+=size
		if e != nil {
			fmt.Printf("Error downloading piece: %v\n", e)
			return
		}


        fmt.Printf("Piece %d downloaded: %d bytes. Remaining bytes: %d\n",i,len(piece),metaInfo.Info.Length-totalDownloaded)
        sha:=sha1.Sum(piece)
        shaD:=hex.EncodeToString(sha[:])
        if p != shaD{
            println("Error: downloaded piece",i,"hash is not the same as actual")
            fmt.Printf("Expected:\n%s\nActual:\n%s\n",p,shaD)
            return
        }

        newFile.Write(piece)
        newFile.Close()
	}



	println("File", metaInfo.Info.Name, "downloaded to", path)

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
func print_info(metaInfo types.MetaInfo) {
	hash, _ := protocol.CalculateInfoHash(metaInfo.Info)
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
