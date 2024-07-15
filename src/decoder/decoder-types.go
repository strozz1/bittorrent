package decoder

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"log"
	"os"

	"github.com/jackpal/bencode-go"
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
	lengthD:= infoMap["length"]
    if lengthD == nil{
        //todo multi file
        return MetaInfo{},errors.New("NO multifile decoder")
    }        
    length:=lengthD.(int)
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

func MetaInfoFromFile(path string) (MetaInfo,[]byte, error) {

	content, e := os.ReadFile(path)
	if e != nil {
		log.Println("Error reading file. ", e)
		os.Exit(1)
	}
	decoded_bytes, e := Decode(content)
	if e != nil {
		return MetaInfo{},nil, e
	}
	dict, _ := decoded_bytes.(map[string]any)
    metaInfo,e:=GetMetaInfo(dict)
	if e != nil {
		return MetaInfo{}, nil,e
	}
    hash_encoded, _ := CalculateInfoHash(metaInfo.Info)
	hash, _ := hex.DecodeString(hash_encoded)
    return metaInfo,hash,nil
}

// Función para calcular el info hash
func CalculateInfoHash(info Info) (string, error) {
	var buffer bytes.Buffer

	err := bencode.Marshal(&buffer, info)
	if err != nil {
		return "", err
	}

	hash := sha1.Sum(buffer.Bytes())

	// Convertir el hash a una cadena hexadecimal
	return hex.EncodeToString(hash[:]), nil
}
