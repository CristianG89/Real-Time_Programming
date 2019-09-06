//https://astaxie.gitbooks.io/build-web-application-with-golang/en/08.1.html

package main

import (
    "fmt"		//For print functions
	"strings"
	"net"		//Interface for network I/O, including TCP/IP, UDP, name resolution and Unix sockets
	"os"		//Interface to operating system functionality
	"runtime"	//For defining cores to use
	"regexp"	//Implements regular expression search
)

const IP_port = "30000"		//Port where server IP is sent
const comm_port = "20014"	//Port for normal communication with server
var server_IP string		//Variables are created by "var" keyword, then name, type and finally a value (this can be skipped)

func checkError(err error) {	//Prints any detected error
    if err != nil {
		fmt.Println("EXIT DUE TO ERROR -> ", err.Error())
		os.Exit(1)
    }
}

func get_UDPAddr(IP_num string, port_num string) (*net.UDPAddr) {	//It returns the IP + port in the as UDPconn object
	service := net.JoinHostPort(IP_num, port_num)				//String with the format "IP:port"
	
	udpAddr, err := net.ResolveUDPAddr("udp4", service)			//UDPconn object is created
	checkError(err)
	
	return udpAddr
}

func send_UDP(write_done chan bool) {		//Client sends a message to the server through UDP protocol
    udpAddr := get_UDPAddr(server_IP, comm_port)	//IP + port in correct format (UDPconn object)

	conn, err := net.DialUDP("udp4", nil, udpAddr)	//Socket is opened to comm through UDP protocol
    checkError(err)
    
    message := []byte("ANYTHING")
    _, err = conn.Write(message)			//Data is sent to server by means of the socket
    checkError(err)

    fmt.Println("I say: " + string(message))

    conn.Close()			//Socket is closed
    write_done <- true		//Channel informs sending has finished
}

func receive_UDP(receive_done chan bool) {	//Client receives a message from the server through UDP protocol
    service := net.JoinHostPort("", comm_port)		//IP + port in correct format (it is just a string here!)
	conn, err := net.ListenPacket("udp4", service)	//Socket opened to listen incoming msg (DialUDP() [UDPconn object] does not work here!)
    checkError(err)

    message := make([]byte, 1024)
    conn.ReadFrom(message[0:])		//Incoming message is read (Read() [UDPconn object] does not work with UDP!)
    checkError(err)
	
    fmt.Println(string(message[0:]))	//Received data has the structure "You said: ..."

	conn.Close()			//Socket is closed
    receive_done <- true	//Channel informs receiving has finished
}

func find_serverIP(IP_found chan bool) {	//Client finds out the server IP
	udpAddr := get_UDPAddr("", IP_port)		//":30000"
	listener, err := net.ListenUDP("udp4", udpAddr)		//An socket is created to listen in all IPs, port 30000
	checkError(err)
	
	message := make([]byte, 1024)
    length, err := listener.Read(message[0:])	//Any incoming message is read
    checkError(err)
	
	server_IP = string(message[0:length])		//Complete answer from the server into string
	re := regexp.MustCompile("[0-9.]")	//Just consider: 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, .
	server_IP = strings.Join(re.FindAllString(server_IP, -1), "")	//Match and join previous characters into string
	
	listener.Close()		//Socket is closed
	IP_found <- true		//Channel informs server IP has been found
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())	//Code is parallelized with as many cores as available
	
	IP_found := make(chan bool, 1)
	write_done := make(chan bool, 1)
	receive_done := make(chan bool, 1)

	go find_serverIP(IP_found)
	<-IP_found
	fmt.Println("Server IP is:", server_IP)

	go send_UDP(write_done)
	<-write_done

	go receive_UDP(receive_done)
	<-receive_done
	
    os.Exit(0)
}