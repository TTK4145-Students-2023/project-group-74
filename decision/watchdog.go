package orderAssigner

import (
	"ElevatorProj/config"
	"ElevatorProj/src/types"
	"time"
)

//******* Constants *******//

const ORDER_WATCHDOG_POLL_RATE = config.ORDER_WATCHDOG_POLL_RATE_MS * time.Millisecond


//******* Functions *******//

func orderWatchdog(
	orderActivatedChn   <-chan   types.OrderType, 
	orderDeactivatedChn <-chan   types.OrderType, 
	orderTimedOutChn      chan<- types.OrderType) {

	var orderTimeouts [types.NUM_FLOORS][types.NUM_BUTTONS]time.Time
	var zeroTime             = time.Time{}
	pollOrderTimeoutsTicker := time.NewTicker(ORDER_WATCHDOG_POLL_RATE)
	
	for {
		select {
		case order := <-orderActivatedChn:
			timeout := orderTimeouts[order.Floor][order.Button]
			if timeout.IsZero() {
				orderTimeouts[order.Floor][order.Button] = time.Now().Add(ORDER_TIMEOUT_PERIOD)
			}

		case order := <-orderDeactivatedChn:
			orderTimeouts[order.Floor][order.Button] = zeroTime

		case <-pollOrderTimeoutsTicker.C:
			for floor := 0; floor < types.NUM_FLOORS; floor++ {
				for button := 0; button < types.NUM_BUTTONS; button++ {
					timeout := orderTimeouts[floor][button]

					if !timeout.IsZero() && timeout.Before(time.Now()) {
						orderTimeouts[floor][button] = zeroTime
						var order types.OrderType
						order.Floor  = floor
						order.Button = types.ButtonType(button)
						
						orderTimedOutChn <- order
					}
				}
			}
		}
	}
}
