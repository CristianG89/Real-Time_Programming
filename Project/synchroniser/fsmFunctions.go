package synchroniser

import (
	"../global"
	"log"
)

// Check if there is an order above a given floor
func (q *queue) isOrdersAbove(thisFloor int) bool {
	for floor := thisFloor + 1; floor < global.NumFloors; floor++ {
		for button := 0; button < global.NumButtons; button++ {
			if q.isOrder(floor, button) {
				return true
			}
		}
	}
	return false
}

// Check if there is an order below a given Floor
func (q *queue) isOrdersBelow(thisFloor int) bool {
	for floor := 0; floor < thisFloor; floor++ {
		for button := 0; button < global.NumButtons; button++ {
			if q.isOrder(floor, button) {
				return true
			}
		}
	}
	return false
}

// Returns the direction the elevator should move to.
func (q *queue) decideDirection(thisFloor, thisDirection int) int {
	
	//If queue is empty, don't move
	if q.isEmpty() {
		return global.DirStop
	}

	switch thisDirection {
	// ...for an elevator moving up
	case global.DirUp:
		if thisFloor < global.NumFloors-1 && q.isOrdersAbove(thisFloor) {
			return global.DirUp
		} else {
			return global.DirDown
		}
	// ...for an elevator moving down
	case global.DirDown:
		if thisFloor > 0 && q.isOrdersBelow(thisFloor) {
			return global.DirDown
		} else {
			return global.DirUp
	}
	// ...elevator is stationary
	case global.DirStop:
		if thisFloor < global.NumFloors-1 && q.isOrdersAbove(thisFloor) {
			return global.DirUp
		} else if thisFloor > 0 && q.isOrdersBelow(thisFloor) {
			return global.DirDown
		} else {
			return global.DirStop
		}
	// ...default, so something wrong happened
	default:
		global.CloseConnectionChan <- true
		global.Restart.Run()
		log.Println(global.ColorR,"Error in decideDirection() with direction: ",thisDirection, "... Returning stop.",global.ColorR, "\n")
		return 0
	}
}

// For FSM. Decides if lift should stop and open door at a given floor and given direction
func (q *queue) checkIfStop(floor, dir int) bool {
	switch dir {

	// ...for an elevator moving up
	case global.DirUp:
		return floor == global.NumFloors-1 || 
			q.isOrder(floor, global.BtnInside) ||
			q.isOrder(floor, global.BtnOutUp) ||
			!q.isOrdersAbove(floor) 

	// ...for an elevator moving down
	case global.DirDown:
		return floor == 0 ||
			q.isOrder(floor, global.BtnInside) ||
			q.isOrder(floor, global.BtnOutDown) ||
			!q.isOrdersBelow(floor)

	// ...for an elevator that is stationary (still want doors to open)
	case global.DirStop:
		return q.isOrder(floor, global.BtnInside) ||
		q.isOrder(floor, global.BtnOutUp) ||
		q.isOrder(floor, global.BtnOutDown) 

	// ...default, so something wrong happened
	default:
		global.CloseConnectionChan <- true
		global.Restart.Run()
		log.Println(global.ColorR, "Error in checkIfStop: can't find direction.", global.ColorN)
	}
	return false
}