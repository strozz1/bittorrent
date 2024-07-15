package protocol

import (
	"fmt"
	"net"
	"strconv"
)



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



