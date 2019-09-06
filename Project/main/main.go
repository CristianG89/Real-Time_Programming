package main

import (
	"../global"
	"../stateMachine"
	"../hw"
	"../synchroniser"
	"../network"
	"log"
	"time"
	"os"
	"os/signal"
)

var numLiftsOnline int 		//Total number of Lifts available in the Network
var arrayOnlineLifts = make(map[string]network.UdpConnection)

var ch_deadLift = make(chan network.UdpConnection)
var ch_moveCost = make(chan global.Message)
var ch_outMsg 	= make(chan global.Message, 10)
var ch_inMsg 	= make(chan global.Message, 10)

var ch_FSM = global.FSM_channels{		//Channels to handle the FSM and outgoing UDP messages
	NewOrder:    	make(chan bool),
	FloorReached: 	make(chan int),
	DoorTimeout:	make(chan bool),
	MotorDir:     	make(chan int, 10),
	FloorLamp:    	make(chan int, 10),
	DoorTimerReset:	make(chan bool),
	DoorLamp:     	make(chan bool, 10),
	OutgoingMsg:  	ch_outMsg,
}

func main() {
	floor, err := hw.HW_Init()	//All lamps are off and lift moves down until the next correct floor
	if err != nil {
		global.Restart.Run()
		log.Fatal(err)
	}

	stateMachine.FSM_Init(ch_FSM, floor)			//FSM initialization, telling it which floor number is the current one
	network.Network_Init(ch_outMsg, ch_inMsg)		//Network initialization, with channels for outcoming and incoming messages as channels
	synchroniser.Init(ch_FSM.NewOrder, ch_outMsg)	//Queue initialization, with new order and outcoming message as channels

	go synchroniser.RunAssign(ch_moveCost, &numLiftsOnline)	//Finds the best lift to assign the next operation
	go manage_Events()	//Manages all events related with hardware (buttons and lamps) as well as incoming UDP messages
	go buttonLamps()	//Manages all buttons lamps (external and internal)
	go securitySTOP()	//Stops the motor if user terminates the program

	infiniteLoop := make(chan bool)
	<-infiniteLoop		//It waits for a channel that will never be 1, so infinite loop (if main function terminates, the whole program stops)
}

//It manages all events happening in the system (so the main loop)
func manage_Events() {
	buttonChan := checkAllButtons()			//It calls an infinite loop to check which buttons have been pushed
	floorChan := checkAllFloors()			//It calls an infinite loop to check which floor the lift is now

	for {
		select {
		case keys := <-buttonChan:			//If any button has been pushed
			switch keys.Button {
			case global.BtnInside:			//More specifically, if one of inside buttons was pushed, then local order requested
				synchroniser.AddLocalOrder(keys.Floor, keys.Button)
			case global.BtnOutUp, global.BtnOutDown:	//Only in case an external button was pushed, the order is sent out by UPD message
				if network.UDP_Alive() {		//In case UDP si alive, message includes the order, button and floor
					ch_outMsg <- global.Message{Category: global.Msg_NewOrder, Floor: keys.Floor, Button: keys.Button}
				}
			}
		case new_Floor := <-floorChan:			//In case a new floor is reached, the proper channel is updated
			ch_FSM.FloorReached <- new_Floor
		case message := <-ch_inMsg:				//In case a new external UDP message is received, it is read to define new actions
			manage_Message(message)
		case connection := <-ch_deadLift:		//In case another lift dies, it is removed from lifts list (and its orders are reassigned)
			handleDeadLift(connection.Addr)
		case order := <-synchroniser.OrderTimeoutChan:		//In case a remote order is not finished after 30 seconds
			log.Println(global.ColorR, "Order timeout, it will be done locally", global.ColorN)
			synchroniser.RemoveRemoteOrdersAt(order.Floor)
			synchroniser.AddRemoteOrder(order.Floor, order.Button, global.Laddr)
		case direction := <-ch_FSM.MotorDir: 	//Defines which direction the motor should go to (or just stop)
			hw.SetMotorDir(direction)
		case curr_floor := <-ch_FSM.FloorLamp:	//Manages which external floor lamp to light
			hw.SetFloorLamp(curr_floor)
		case value := <-ch_FSM.DoorLamp:		//Manages the door open lamp
			hw.SetDoorLamp(value)
		}
	}
}

// It continuously checks which buttons (external and internal) have been pushed
func checkAllButtons() <-chan global.Keypress {
	ch_button := make(chan global.Keypress)
	go func() {		//This structure allows the function to return something with a infinite loop inside...
		var buttonState [global.NumFloors][global.NumButtons]bool

		for {
			for floor := 0; floor < global.NumFloors; floor++ {				//Four floors in total!
				for button := 0; button < global.NumButtons; button++ {		//Three buttons per floor!
					if (floor == 0 && button == global.BtnOutDown) ||
						(floor == global.NumFloors-1 && button == global.BtnOutUp) {
						continue
					}
					if hw.ReadButton(floor, button) {
						if !buttonState[floor][button] {
							ch_button <- global.Keypress{Button: button, Floor: floor}
						}
						buttonState[floor][button] = true
					} else {
						buttonState[floor][button] = false
					}
				}
			}
			time.Sleep(time.Millisecond)
		}
	}()
	return ch_button
}

// It checks the queues and updates all button lamps accordingly
func buttonLamps() {	// This function cannot be included in "manage_Events" because it includes a for loop, that takes indeed some time...
	for {
		<-global.SyncLightsChan
		for floor := 0; floor < global.NumFloors; floor++ {				//Four floors in total!
			for button := 0; button < global.NumButtons; button++ {		//Three buttons per floor!
				if (button == global.BtnOutUp && floor == global.NumFloors-1) || (button == global.BtnOutDown && floor == 0) {
					continue
				} else {
					switch button {
					case global.BtnInside:
						hw.SetButtonLamp(floor, button, synchroniser.IsLocalOrder(floor, button))
					case global.BtnOutUp, global.BtnOutDown:
						hw.SetButtonLamp(floor, button, synchroniser.IsRemoteOrder(floor, button))
					}
				}
			}
		}
	}
}

// It continuously checks is which floor the local lift is now
func checkAllFloors() <-chan int {
	ch_floor := make(chan int)		//This structure allows the function to return something with a infinite loop inside...
	go func() {
		prevFloor := hw.Floor()
		for {
			newFloor := hw.Floor()
			if newFloor != prevFloor && newFloor != -1 {
				ch_floor <- newFloor
			}
			prevFloor = newFloor
			time.Sleep(time.Millisecond)
		}
	}()
	return ch_floor
}

// It handles incoming messages from network (if external lift is still alive, new order [and its cost] or already complete orders)
func manage_Message(msg global.Message) {
	const aliveTimeout = 2 * time.Second

	switch msg.Category {
	case global.Msg_Alive:
		if connection, exist := arrayOnlineLifts[msg.Addr]; exist {
			connection.Timer.Reset(aliveTimeout)
		} else {
			newConnection := network.UdpConnection{msg.Addr, time.NewTimer(aliveTimeout)}
			arrayOnlineLifts[msg.Addr] = newConnection
			numLiftsOnline = len(arrayOnlineLifts)
			go UDP_connect_Timer(&newConnection)
			log.Printf("%sLift with IP %s joined!%s", global.ColorG, msg.Addr[0:15], global.ColorN)
		}
	case global.Msg_NewOrder:
		cost := synchroniser.CalculateCost(msg.Floor, msg.Button, stateMachine.Floor, hw.Floor(), stateMachine.Dir)
		ch_outMsg <- global.Message{Category: global.Msg_Cost, Floor: msg.Floor, Button: msg.Button, Cost: cost}
	case global.Msg_CompleteOrder:
		synchroniser.RemoveRemoteOrdersAt(msg.Floor)
	case global.Msg_Cost:
		ch_moveCost <- msg
	}
}

// It checks if a lifts has been lost from network
func UDP_connect_Timer(connection *network.UdpConnection) {
	<-connection.Timer.C
	ch_deadLift <- *connection
}

// Local lift has been removed, so it tries to connect back to the UDP network
func UDP_Reconnection() {
	network.UDP_Reconnect()
}

// It removes a dead lift from the list of online lifts, and reassigns any order assigned to it
func handleDeadLift(deadAddr string) {
	log.Printf("%sLift with IP %s left!%s", global.ColorR, deadAddr[0:15], global.ColorN)
	delete(arrayOnlineLifts, deadAddr)
	numLiftsOnline = len(arrayOnlineLifts)
	synchroniser.ReassignOrders(deadAddr, ch_outMsg)		//Pending orders are reassigned
	go UDP_Reconnection()		// Handles reconnection to UDP network
}

// It switches the motor off if the program is terminated with CTRL+C
func securitySTOP() {
	var ctrl_c = make(chan os.Signal)		//Channel of type "Signal"
	signal.Notify(ctrl_c, os.Interrupt)		//Notifies the "Signal" when os.Interrupt occurs
	<-ctrl_c
	hw.SetMotorDir(global.DirStop)			//As it leaves, motor must be stopped
	log.Fatal(global.ColorR, "Program terminated manually", global.ColorN)
}