//https://play.golang.org/

package main

import (
    "fmt"		//For print functions
	"sync"		//For mutex functions
    "runtime"	//For defining cores to use
)

var mutex = &sync.Mutex{}	//Mutex object is created
var i = 0					//Global variable

func inc(done_1 chan bool) {
	mutex.Lock()	//Access to var i is now locked (other threads must wait to modify var i)
	for j:=0; j<1000000; j++ {
		i++
	}
	mutex.Unlock()	//Access to var i is now unlocked (other threads can start to modify var i)
    done_1 <- true	//true is sent to the channel to inform other threads, this one has finished
}

func dec(done_2 chan bool) {
	mutex.Lock()
	for j:=0; j<1000000; j++ {
		i--
	}
	mutex.Unlock()
    done_2 <- true
}

func main() {
    //GOMAXPROCS defines how many cores of the PC should be used to run the different threads (it saves time)
	runtime.GOMAXPROCS(runtime.NumCPU())	//NumCPU() returns the number of cores of this PC
	//fmt.Printf("cores: %d\n", runtime.NumCPU())		//The result is just 1!!??
	
	done_1 := make(chan bool, 1)	//Channel created. It is boolean, so its size is 1
	done_2 := make(chan bool, 1)
	
	go inc(done_1)	//Functions inc and dec are executed by another thread (goroutine).
	go dec(done_2)	//Channel flags are the arguments
	
	//Just for synchronization purposes, not mutual exclusion!
	<-done_1	//Main thread waits until both channels are true (so both functions have finished)
	<-done_2	//(on the left side a variable could have been written to save current channel status in it)
	
    fmt.Printf("The final value is: %d\n", i)
}