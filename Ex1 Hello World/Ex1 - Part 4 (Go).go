//https://play.golang.org/

package main

import (
    "fmt"		//For print functions
    "time"		//For delay functions
    "runtime"	//For defining cores to use
)
var i = 0

func inc() {
	for j:=0; j<1000000; j++ {
		i++
	}
}

func dec() {
	for j:=0; j<1000000; j++ {
		i--
	}
}

func main() {
    //GOMAXPROCS defines how many cores of the PC should be used to run the different threads (it saves time)
	runtime.GOMAXPROCS(runtime.NumCPU())	//NumCPU() returns the number of cores of this PC
	//fmt.Printf("cores: %d\n", runtime.NumCPU())		//The result is just 1!!??

	go inc()	//Functions inc and dec are executed by new threads (goroutine)
	go dec()
	
    //We have no way (so far) to wait for the completion of a goroutine. For now: Sleep.
    time.Sleep(100*time.Millisecond)
	
    fmt.Printf("The final value is: %d\n", i)
}