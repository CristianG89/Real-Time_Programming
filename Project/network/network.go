package network

import (
	"../global"
	"encoding/json"
	"strings"
	"strconv"
	"log"
	"net"
	"time"
)


const localPort = 31245			//Local and broadcast ports
const broadcastPort = 31245		//(chosen randomly to reduce probability of port collision)
const messageSize = 1024						//Size of all UDP messages
const alive_delay = 400 * time.Millisecond		//Time between life signal messages

//UDPAddr is a predefined struct having: IP ([]byte), Port (int), Zone (string)
var broadcast_addr *net.UDPAddr 	//Broadcast address
var local_addr *net.UDPAddr 		//Local address

var udpSend = make(chan udpMessage)
var udpReceive = make(chan udpMessage, 10)

type UdpConnection struct {
	Addr  string
	Timer *time.Timer
}

type udpMessage struct {
	raddr  string 	//if receiving raddr=senders address, if sending raddr should be set to "broadcast" or an ip:port
	data   []byte
	length int 		//length of received data, in #bytes // N/A for sending
}

func UDP_Init() (err error) {
	broadcast_addr, err = net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(broadcastPort))	//Generating broadcast address
	if err != nil {return err}

	tempConn, err := net.DialUDP("udp4", nil, broadcast_addr)		//Generating localocal_address
	if err != nil {return err}
	defer tempConn.Close()
	local_addr, err = net.ResolveUDPAddr("udp4", tempConn.LocalAddr().String())
	if err != nil {return err}
	local_addr.Port = localPort
	global.Laddr = local_addr.String()

	localocalConn, err := net.ListenUDP("udp4", local_addr)				//Local listener created
	if err != nil {return err}

	broadcastConn, err := net.ListenUDP("udp4", broadcast_addr)		//Broadcast listener created
	if err != nil {
		localocalConn.Close()
		return err
	}

	go UDP_Receive(localocalConn, broadcastConn)
	go UDP_Transmit(localocalConn, broadcastConn)
	go UDP_Close(localocalConn, broadcastConn)

	//	log.Printf("Generating local address: \t Network(): %s \t String(): %s \n", local_addr.Network(), local_addr.String())
	//	log.Printf("Generating broadcast address: \t Network(): %s \t String(): %s \n", broadcast_addr.Network(), broadcast_addr.String())
	return err
}

// Continuously transmits data through UDP protocol
func UDP_Transmit(localConn, broadConn *net.UDPConn) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(global.ColorR, "ERROR: UDP_Transmit - Connec. closed.", global.ColorN)
			localConn.Close()
			broadConn.Close()
		}
	}()

	var err error 		//Variables cannot be defined as := inside a for loop!!
	var length int

	for {
		msg := <-udpSend		//Infinite loop here stuck until udpSend changes its value
		if msg.raddr == "broadcast" {
			length, err = localConn.WriteToUDP(msg.data, broadcast_addr)
		} else {
			raddr, err := net.ResolveUDPAddr("udp", msg.raddr)
			if err != nil {panic(err)}					//Extra panic necessary because the result is used later...
			length, err = localConn.WriteToUDP(msg.data, raddr)
		}
		if err != nil || length < 0 {panic(err)}		//If there was an error during writing operation...
	}
}

// Updates any data received (either from local or boradcast addresses)
func UDP_Receive(localConn, broadConn *net.UDPConn) {
	defer func() {
		if r := recover(); r != nil {		//It seems this function is never executed...
			log.Println(global.ColorR, "ERROR: UDP_Receive - Connec. closed.", global.ColorN)
			localConn.Close()
			broadConn.Close()
		}
	}()

	broadConn_rcv_ch := make(chan udpMessage)
	localConn_rcv_ch := make(chan udpMessage)

	go UDP_Reader(localConn, localConn_rcv_ch)		//Read incoming messages from local address
	go UDP_Reader(broadConn, broadConn_rcv_ch)		//Read incoming messages from remote addresses

	for {
		select {
		case a := <-broadConn_rcv_ch:
			udpReceive <- a
		case b := <-localConn_rcv_ch:
			udpReceive <- b
		}
	}
}

// Continuously reads data through UDP protocol from local and broadcast addresses
func UDP_Reader(conn *net.UDPConn, rcv_ch chan<- udpMessage) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(global.ColorR, "ERROR: UDP_Reader", conn.LocalAddr().String(),"- Connec. closed.", global.ColorN)
			conn.Close()
		}
	}()

	for {		//Infinite loop
		buf := make([]byte, messageSize)		//The buffer NEEDS to be defined inside the for loop...
		length, raddr, err := conn.ReadFromUDP(buf)		//Loop here stuck until it reads something
		if ((err != nil) || (length < 0)) {panic(err)}		//If there was an error during writing operation...

		rcv_ch <- udpMessage{raddr: raddr.String(), data: buf, length: length}
	}
}

// As a security, another dedicated function to close all ports if necessary...
func UDP_Close(localConn, broadConn *net.UDPConn) {
	<-global.CloseConnectionChan
	localConn.Close()
	broadConn.Close()
}

func Network_Init(outgoingMsg, incomingMsg chan global.Message) {
	err := UDP_Init()
	if err != nil {
		log.Println(global.ColorR, "UDP_Init() error. Stand-alone mode", global.ColorN)
		go UDP_Reconnect()
	} else {
		log.Println(global.ColorG, "Network initialised.", global.ColorN)
	}

	go LifeSignal(outgoingMsg)
	go checkOutgoingMsg(outgoingMsg)
	go checkIncomingMsg(incomingMsg)
}

// Checks continuously to reconnect when UDP communication is lost
func UDP_Reconnect() {
	for {		//Infinite loop until UDP comm. recovered
		err := UDP_Init()
		if err == nil {break}
	}
	log.Println(global.ColorG, "Network recovered.", global.ColorN)
}

//Checks if UDP comm. is alive before sending any outgoing message
func UDP_Alive() bool {
	err := UDP_Init()
	if err == nil {return true}
	stringArray := strings.Split(err.Error(), " ")
	txt := stringArray[len(stringArray)-1]		//Only the last word of the error is saved

	if txt == "unreachable" {	//If address is not reachadable
		return false
	} else {
		return true
	}
}

//Sends messages periodically on network to notify all lifts that this lift is still online..
func LifeSignal(outgoingMsg chan<- global.Message) {
	aliveMsg := global.Message{Category: global.Msg_Alive, Floor: -1, Button: -1, Cost: -1}
	for {
		outgoingMsg <- aliveMsg
		time.Sleep(alive_delay)
	}
}

//Checks continuously for sending new messages on network (by reading channel OutgoingMsg). Each message is sent by UDP protocol as JSON.
func checkOutgoingMsg(outgoingMsg <-chan global.Message) {
	for {
		msg := <-outgoingMsg	//Infinite loop, here stuck until new message request comes

		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("%sjson.Marshal error: %v\n%s", global.ColorR, err, global.ColorN)
		}
		udpSend <- udpMessage{raddr: "broadcast", data: jsonMsg, length: len(jsonMsg)}
	}
}

//Checks continuously for incoming messages from network (by reading channel udpReceive). Each message is read as JSON.
func checkIncomingMsg(incomingMsg chan<- global.Message) {
	for {
		udpMessage := <-udpReceive
		var message global.Message

		if err := json.Unmarshal(udpMessage.data[:udpMessage.length], &message); err != nil {
			log.Printf("%sjson.Unmarshal error: %v\n%s", global.ColorR, err, global.ColorN)
		}
		message.Addr = udpMessage.raddr
		incomingMsg <- message
	}
}