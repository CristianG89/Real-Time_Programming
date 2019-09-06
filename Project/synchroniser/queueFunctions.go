package synchroniser

import (
	"../global"
	"time"
	"log"
)

const orderTimeout = 30 * time.Second

// Sets a new order if order is not in queue, synch lamps and creates backup
func (q *queue) setOrder(floor, button int, status orderStatus) {
	if q.isOrder(floor, button) == status.active {
		return 	//Leave because order was already there
	}
	q.matrix[floor][button] = status
	global.SyncLightsChan <- true
	takeBackup <- true		//Save a Backup
}

// Check if there is an order in a specific floor and button 
func (q *queue) isOrder(floor, button int) bool {
	return q.matrix[floor][button].active
}

// Iterates through a queue to see if its empty
func (q *queue) isEmpty() bool {
	for floor := 0; floor < global.NumFloors; floor++ {
		for button := 0; button < global.NumButtons; button++ {
			if q.matrix[floor][button].active { 	//Order found
				return false
			}
		}
	}
	return true
}

// Makes a new queue and copies all elements to this queue
func (q *queue) copyQueue() *queue {
	queueCopy := new(queue)
	for floor := 0; floor < global.NumFloors; floor++ {
		for button := 0; button < global.NumButtons; button++ {
			queueCopy.matrix[floor][button] = q.matrix[floor][button]
		}
	}
	return queueCopy
}

// Starts a timer for 30 seconds for an order
func (q *queue) startTimer(floor, button int) {
	q.matrix[floor][button].timer = time.NewTimer(orderTimeout)
	<-q.matrix[floor][button].timer.C
	OrderTimeoutChan <- global.Keypress{Button: button, Floor: floor}
}

// Stops the timer
func (q *queue) stopTimer(floor, button int) {
	if q.matrix[floor][button].timer != nil {
		q.matrix[floor][button].timer.Stop()
	}
}

/* Calculates the cost of a requested order for a given lift, with the following pattern:
   - 2 points for each travel between adjacent floors
   - 1 point if the elevator starts between two floors
   - 2 points if lift already moving in wrong direction (high cost --> low priority
*/
func CalculateCost(targetFloor, targetButton, prevFloor, currFloor, currDir int) int {
	q := local.copyQueue()
	q.setOrder(targetFloor, global.BtnInside, orderStatus{true, "", nil})

	cost := 0
	floor := prevFloor
	dir := currDir

	// Between floors
	if currFloor == -1 {
		cost++

	// Moving
	} else if dir != global.DirStop {
		cost += 2
	}
	//floor, dir = incrementFloor(floor, dir)
	if dir == global.DirDown {
		floor--
	} else if dir == global.DirUp{
		floor++
	} else if dir != global.DirStop {
		global.CloseConnectionChan <- true
		global.Restart.Run()
		log.Fatalln(global.ColorR, "ERROR in incrementFloor: cant find direction.", global.ColorN)
	}

	// Simulate the rest of the trip to find cost for an elevator to reach its destination.
	// (only up to 10 iterations in order to prevent possible infinite loop)
	for n := 0; !(floor == targetFloor && q.checkIfStop(floor, dir)) && n < 10; n++ {
		if q.checkIfStop(floor, dir) {
			cost += 2
			q.setOrder(floor, global.BtnOutUp, inactive)
			q.setOrder(floor, global.BtnOutDown, inactive)
			q.setOrder(floor, global.BtnInside, inactive)
		}
		dir = q.decideDirection(floor, dir)
		//floor, dir = incrementFloor(floor, dir)
		if dir == global.DirDown {
			floor--
		} else if dir == global.DirUp{
			floor++
		} else if dir != global.DirStop {
			global.CloseConnectionChan <- true
			global.Restart.Run()
			log.Fatalln(global.ColorR, "ERROR in incrementFloor: cant find direction.", global.ColorN)
		}
		cost += 2
	}
	return cost
}