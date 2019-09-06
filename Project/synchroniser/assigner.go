
package synchroniser

import (
	"../global"
	"log"
	"time"
)

type reply struct {
	cost int
	lift string
}
type order struct {
	floor  int
	button int
	timer  *time.Timer
}

// Always keeps track of which elevator is best for each order
func RunAssign(costReply <-chan global.Message, numLiftsOnline *int) {
	
	unassigned := make(map[order][]reply)
	var timeout = make(chan *order)
	const timeoutDuration = 10 * time.Second

	for {
		select {
		case message := <-costReply:
			newOrder := order{floor: message.Floor, button: message.Button}
			newReply := reply{cost: message.Cost, lift: message.Addr}

			for oldOrder := range unassigned {
				if oldOrder.floor == newOrder.floor && oldOrder.button == newOrder.button {
					newOrder = oldOrder
				}
			}

			// Check if order in queue.
			if replyList, exist := unassigned[newOrder]; exist {
				// Check if newReply already is registered.
				found := false
				for _, reply := range replyList {
					if reply == newReply {
						found = true
					}
				}
				// Add it if it wasn't.
				if !found {
					unassigned[newOrder] = append(unassigned[newOrder], newReply)
					newOrder.timer.Reset(timeoutDuration)
				}
			} else {
				// If order not in queue at all, init order list with it
				newOrder.timer = time.NewTimer(timeoutDuration)
				unassigned[newOrder] = []reply{newReply}
				go costTimer(&newOrder, timeout)
			}
			chooseBestLift(unassigned, numLiftsOnline, false)

		case <-timeout:
			log.Println(global.ColorR, "Not all costs received in time!", global.ColorN)
			chooseBestLift(unassigned, numLiftsOnline, true)
		}
	}
}

/* Checks if any order is waiting for a lift assignment to collect enough info to assign.
   For all orders that have, it selects a lift and adds it to the queue.
   It assumes all lifts always make the same decision, but if they do not,
   a timer for each order assured that this never gives unhandled orders. */
func chooseBestLift(unassigned map[order][]reply, numLiftsOnline *int, orderTimedOut bool) {
	const maxInt = int(^uint(0) >> 1)
	// Loop through all lists.
	for order, replyList := range unassigned {
		// Check if the list is complete or the timer has timed out.
		if len(replyList) == *numLiftsOnline || orderTimedOut {
			lowestCost := maxInt
			var bestLift string

			// Loop through costs in each complete list.
			for _, reply := range replyList {
				if reply.cost < lowestCost {
					lowestCost = reply.cost
					bestLift = reply.lift
				} else if reply.cost == lowestCost {
					// Prioritise on lowest IP value if cost is the same.
					if reply.lift < bestLift {
						lowestCost = reply.cost
						bestLift = reply.lift
					}
				}
			}
			AddRemoteOrder(order.floor, order.button, bestLift)
			order.timer.Stop()
			delete(unassigned, order)
		}
	}
}

// Time out to accomplish the requested order
func costTimer(newOrder *order, timeout chan<- *order) {
	<-newOrder.timer.C
	timeout <- newOrder
}