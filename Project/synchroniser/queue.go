package synchroniser

import (
	"../global"
	"../network"
	"log"
	"time"
)

//Defines an inactive orderStatus
var inactive = orderStatus{active: false, addr: "", timer: nil}

var local queue 	//Local queue
var remote queue 	//Remote queue

var updtLocalQueue = make(chan bool)
var takeBackup = make(chan bool, 10)
var OrderTimeoutChan = make(chan global.Keypress)
var newOrderQueue chan bool

// Defines an orderStatus, which consists on info of whether it is active or not,
// which lift has the order, and timer to keep track of how long the order is active
type orderStatus struct {
	active bool
	addr   string      `json:"-"`
	timer  *time.Timer `json:"-"`
}

// Defines a queue, as a 2D matrix of orderStatus. Based on floor and button
type queue struct {
	matrix [global.NumFloors][global.NumButtons]orderStatus
}

// Queue initialisation
func Init(newOrderTemp chan bool, outgoingMsg chan global.Message) {
	newOrderQueue = newOrderTemp
	go updateLocalQueue()
	runBackup(outgoingMsg)
	log.Println(global.ColorG, "Queue initialised.", global.ColorN)
}

// setOrder to local queue
func AddLocalOrder(floor int, button int) {
	local.setOrder(floor, button, orderStatus{true, "", nil})
	newOrderQueue <- true
}

// setOrder to remote queue, if order is not already there
func AddRemoteOrder(floor, button int, addr string) {
	alreadyExist := IsRemoteOrder(floor, button)
	remote.setOrder(floor, button, orderStatus{true, addr, nil})
	if !alreadyExist {
		go remote.startTimer(floor, button)
	}
	updtLocalQueue <- true
}

// Deletes oders at a given floor in remote queue, stops timer and set order inactive
func RemoveRemoteOrdersAt(floor int) {
	for button := 0; button < global.NumButtons; button++ {
		remote.stopTimer(floor, button)
		remote.setOrder(floor, button, inactive)
	}
	updtLocalQueue <- true
}

// Delete orders at a given floor for both local and remote
func RemoveOrdersAt(floor int, outgoingMsg chan<- global.Message) {
	for button := 0; button < global.NumButtons; button++ {
		remote.stopTimer(floor, button)
		local.setOrder(floor, button, inactive)
		remote.setOrder(floor, button, inactive)
	}
	if network.UDP_Alive() {  	// Checks that if its actually broadcasting to send the next UDP message
		outgoingMsg <- global.Message{Category: global.Msg_CompleteOrder, Floor: floor}
	}
}

// Takes order from a dead lift, reassign these orders to the network as new orders
func ReassignOrders(deadAddr string, outgoingMsg chan<- global.Message) {
	for floor := 0; floor < global.NumFloors; floor++ {
		for button := 0; button < global.NumButtons; button++ {
			if remote.matrix[floor][button].addr == deadAddr {
				remote.setOrder(floor, button, inactive)
				if network.UDP_Alive() {
					outgoingMsg <- global.Message{Category: global.Msg_NewOrder, Floor: floor, Button: button}
				}
			}
		}
	}
}

// Checks remote queue for new orders assigned to this lift and copies them to local queue
func updateLocalQueue() {
	for {
		<-updtLocalQueue
		for floor := 0; floor < global.NumFloors; floor++ {
			for button := 0; button < global.NumButtons; button++ {
				if remote.isOrder(floor, button) {
					if button != global.BtnInside && remote.matrix[floor][button].addr == global.Laddr {
						if !local.isOrder(floor, button) {
							local.setOrder(floor, button, orderStatus{true, "", nil})
							newOrderQueue <- true
						}
					}
				}
			}
		}
	}
}

// Uses the "isOrder" from queueFunctions to check for local order.
func IsLocalOrder(floor, button int) bool {
	return local.isOrder(floor, button)
}

// Uses the "isOrder" from queueFunctions to check for remote order.
func IsRemoteOrder(floor, button int) bool {
	return remote.isOrder(floor, button)
}

// Uses the "checkIfStop" from queueFunctions to tell if an elevator should stop or not.
func CheckIfStop(floor, dir int) bool {
	return local.checkIfStop(floor, dir)
}

// Uses the "decideDirection" from queueFunctions to tell wich direction it should move.
func ChooseDirection(floor, dir int) int {
	return local.decideDirection(floor, dir)
}