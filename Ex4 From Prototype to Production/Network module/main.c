package main

import (
	"./network/bcast"		//File manually include, with the same name as the last folder
	"./network/localip"		//File manually include, with the same name as the last folder
	"./network/peers"		//File manually include, with the same name as the last folder
	"flag"					//Package flag implements command-line flag parsing
	"fmt"
	"os"
	"time"
)

const (
    peers_port = 15647
	bcast_port = 16569
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")	//StringVar defines a string flag with specified name, default value, and usage string
	flag.Parse()	//Parse analyses/dissects the command line into the defined flags.

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the process ID)
	if id == "" {
		localIP, err := localip.LocalIP()		//This function returns the IP of the present PC
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())	//Getpid() returns the process id of the caller
	}
	
	// We can disable/enable the transmitter after it has been started. This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	//It sends the ID to all IPs (broadcast) by UDP protocol. Infinite loop
	go peers.Transmitter(peers_port, id, peerTxEnable)	//Transmitter(port, id, transmit channel [bool])
	
	// We make a channel for receiving updates on the id's of the peers that are alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)		//Channel of type "PeerUpdate" created (which actually it is a struct)
	go peers.Receiver(peers_port, peerUpdateCh)		//Receiver(port, receive port [PeerUpdate])
	
	// We make channels for sending and receiving our custom data types and start the transmitter/receiver pair on some port
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	
	// These functions can take any number of channels! It is also possible to start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(bcast_port, helloTx)	//Infinite loop. It sends data to all IPs broadcast by UDP protocol
	go bcast.Receiver(bcast_port, helloRx)		//Infinite loop.

	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}		//Struct of type HelloMsg (String, int)
		for {	//Only "for" is infinite loop in Golang
			helloMsg.Iter++		//Every second, the same ID message is sent (through broadcast [above]), but with different Iteration	
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {		//Only "for" is infinite loop in Golang
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
		}
	}
}