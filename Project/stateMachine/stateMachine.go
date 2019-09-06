package stateMachine

import (
	"../global"
	"../synchroniser"
	"log"
	"time"
)

const (					// The 3 states of the lift are:
	idle int = 0		// Idle: Lift is at a door in stationary postion with its door closed
	moving int = 1		// Moving: Lift is moving between floors (or passing at a floor)
	doorOpen int = 2	// Door open: Lift is at a floor with its door open
)

var state int
var Floor int 	//Varaibles starting with capital letters can be used maong packages...
var Dir int

// Function to initialise the FSM
func FSM_Init(ch global.FSM_channels, startFloor int) {
	state = idle //Default values given during initialisation
	Dir = global.DirStop
	Floor = startFloor

	go doorTimer(ch.DoorTimeout, ch.DoorTimerReset)
	go FSM_Run(ch)

	log.Println(global.ColorG, "FSM initialised.", global.ColorN)
}

//This functions runs the corresponding function when a new event occurs
func FSM_Run(ch global.FSM_channels) {
	for {
		select {							// FSM works under 3 possible state transitions:
		case <-ch.NewOrder:					// 1 - New order: An order is added to the queue. 
			eventNewOrder(ch)
		case Floor := <-ch.FloorReached:	// 2 - Floor reached  Lift arrives at a floor
			eventFloorReached(ch, Floor)
		case <-ch.DoorTimeout:				// 3 - Door timeout: Door has been open for sometime, so time to close it
			eventDoorTimeout(ch)
		}
	}
}

func eventNewOrder(ch global.FSM_channels) {
	log.Printf("%sEVENT: %sNew order in state %v.%s", global.ColorY, global.ColorC, check_state(state), global.ColorN)
	switch state {
	case idle:
		// Here it is determined whether the lift should move up, down or remain in stop 
		Dir = synchroniser.ChooseDirection(Floor, Dir)
		// CheckIfStop sees if an order corresponds to the floor the lift is in
		// If it does not have to stop, the lift starts moving in the direction chosen
		// Otherwise, the door opens
		if synchroniser.CheckIfStop(Floor, Dir) {
			ch.DoorTimerReset <- true
			synchroniser.RemoveOrdersAt(Floor, ch.OutgoingMsg)
			ch.DoorLamp <- true
			state = doorOpen
			log.Printf("%sEVENT: %sAlready in floor. Opening door...%s", global.ColorY, global.ColorM, global.ColorN)
		} else {
			ch.MotorDir <- Dir
			state = moving
		}
	case moving:
		// Ignored when the lift is already in motion.
	case doorOpen:
		// If the order corresponds to the same floor where the door is open
		// Then the order is considered serviced and removed
		if synchroniser.CheckIfStop(Floor, Dir) {
			ch.DoorTimerReset <- true
			synchroniser.RemoveOrdersAt(Floor, ch.OutgoingMsg)
		}
	default:
		global.CloseConnectionChan <- true
		global.Restart.Run()
		log.Println(global.ColorR, "ERROR. This state does not exist", global.ColorN)
	}
}

// When lift reaches a floor and an order to stop exists, lifts stops and door is opened
func eventFloorReached(ch global.FSM_channels, newFloor int) {
	log.Printf("%sEVENT: %sFloor %d reached in state %s.%s", global.ColorY, global.ColorB, newFloor+1, check_state(state), global.ColorN)
	Floor = newFloor
	ch.FloorLamp <- Floor
	if (state == doorOpen && Dir == 0) {state = moving}		//This "imposible" situation is just ignored...
	switch state {
	case moving:
		if synchroniser.CheckIfStop(Floor, Dir) {
			ch.DoorTimerReset <- true
			synchroniser.RemoveOrdersAt(Floor, ch.OutgoingMsg)
			ch.DoorLamp <- true
			Dir = global.DirStop
			ch.MotorDir <- Dir
			state = doorOpen
			log.Printf("%sEVENT: %sRequested floor reached. Opening door...%s", global.ColorY, global.ColorM, global.ColorN)
		}
	default:
		global.CloseConnectionChan <- true
		global.Restart.Run()
		log.Fatalf("%sNo sense to arrive at a floor in state %s%s.\n", global.ColorR, check_state(state), global.ColorN)
	}
}

// When door is open for a while in a floor, it is checked if close door (and go idle) or start moving
func eventDoorTimeout(ch global.FSM_channels) {
	switch state {
	case doorOpen:
		ch.DoorLamp <- false
		Dir = synchroniser.ChooseDirection(Floor, Dir)
		ch.MotorDir <- Dir
		if Dir == global.DirStop {
			state = idle
		} else {
			state = moving
		}
		log.Printf("%sEVENT: %s...closing door. Now in state %s.%s", global.ColorY, global.ColorM, check_state(state), global.ColorN)
	default:
		global.CloseConnectionChan <- true
		global.Restart.Run()
		log.Printf(global.ColorR, "No sense to time out when not in state door open", global.ColorN)
	}
}

// It keeps a timer of fixed duration, and notifies to FSM when it times out
func doorTimer(timeout chan<- bool, reset <-chan bool) {
	const doorOpenTime = 3 * time.Second
	timer := time.NewTimer(0)
	timer.Stop()

	for {
		select {
		case <-reset:
			timer.Reset(doorOpenTime)
		case <-timer.C:
			timer.Stop()
			timeout <- true
		}
	}
}

// It returns the current FSM state as a string
func check_state(state int) string {
	switch state {
	case idle:
		return "idle"
	case moving:
		return "moving"
	case doorOpen:
		return "door open"
	default:
		return "error: bad state"
	}
}