package localip

import (
	"net"
	"strings"
)

var localIP string

func LocalIP() (string, error) {		//It returns the IP of the current PC
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]		//Local address taken (IP:port), but string is cut until ":" to take IP
	}
	return localIP, nil
}