//https://astaxie.gitbooks.io/build-web-application-with-golang/en/08.1.html

package main

import (
	"fmt"		//For print functions
	"net"		//Interface for network I/O, including TCP/IP, UDP, name resolution and Unix sockets
	"os"		//Interface to operating system functionality
	"runtime"	//For defining cores to use
	"time"
)

const (
    server_IP = "10.100.23.242"
	client_IP = "10.100.23.151"
    comm_port = "33546"				//This port requires using delimited messages with "\x00" as the marker
	//comm_port = "34933"			//Alternative port. This one requires messages of fixed size 1024
)


func checkError(err error) {    	//Prints any detected error
    if err != nil {
        fmt.Println("EXIT DUE TO ERROR -> ", err.Error())
        os.Exit(1)
    }
}

func get_TCPAddr(IP_num string, port_num string) (*net.TCPAddr) { //It returns the IP + port in the as UDPconn object
    service := net.JoinHostPort(IP_num, port_num)
	
    tcpAddr, err := net.ResolveTCPAddr("tcp4", service)          //UDPconn object is created
    checkError(err)
	
    return tcpAddr
}

func client2server() {						//Client send connection request to the server
    serverAddr := get_TCPAddr(server_IP, comm_port)
    serverConn, err := net.DialTCP("tcp4", nil, serverAddr)		//Connection established from client to server (as UDP)
    checkError(err)
    
    //Unlike UDP, TCP requires to establish communication from server side too, so a request to do so is sent from client
	comm_request := []byte("Connect to:" + net.JoinHostPort(client_IP, comm_port) + "\x00")		//String into an array of bytes
    serverConn.Write(comm_request)
	serverConn.Close()
}

func server2client() (*net.TCPConn) {			//Client waits for the server connection request
	clientAddr := get_TCPAddr(client_IP, comm_port)		//IP "" is also accepted
    tcpListener, err := net.ListenTCP("tcp4", clientAddr)
    checkError(err)
    
    conn, err := tcpListener.AcceptTCP()		//Waits to receive Server TCP request, and accepts it
    checkError(err)
	
	return conn
}

func establish_Conn() (*net.TCPConn) {			//To establish (bidirectional) TCPcommunication with the server
	client2server()
	TCPconn := server2client()			//This function returns the socket to be used from now on in all write/read
	
	fmt.Println("TCP connection established!")

	return TCPconn
}

func TCP_test(conn *net.TCPConn, message string) {
    snd_data := make([]byte, 1024)
    rcv_data := make([]byte, 1024)
    
	snd_data = []byte(message + "\x00")			//String into an array of bytes
    _, err := conn.Write(snd_data)
    checkError(err)

    length, err := conn.Read(rcv_data[0:])		//It seems that Read() function works with TCP (not with UDP)...
	checkError(err)
	
    fmt.Println(string(rcv_data[0:length]))
    time.Sleep(time.Second)						//Delay between messages...
}

func test_Connection(conn *net.TCPConn, test_done chan bool) {
    fmt.Println("TEST STARTS!")

    TCP_test(conn, "Whats up")					//For some reason, not all answers are received (even with delay)
    TCP_test(conn, "Hardwarebeschreibungsprachen")
    TCP_test(conn, "Easy peasy")
    TCP_test(conn, "Fucking CA")
	
	fmt.Println("TEST FINISHED!")
    
    test_done <- true
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())	//Code is parallelized with as many cores as available
	
    test_done := make(chan bool, 1)

	TCPconn := establish_Conn()		//TCP comm established (I cannot make it in parallel because it returns a result)

    go test_Connection(TCPconn, test_done)
    <- test_done		//Continue once this flag is true

    TCPconn.Close()
    os.Exit(0)
}