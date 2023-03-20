package DLOCC

import (
	"encoding/json"
	"fmt"
	"project-group-74/localTypes"
	"os/exec"
	"time"
)

func orderWatchdog(
	orderActivatedChn <-chan localTypes.FOREIGN_ORDER_TYPE,
	orderDeactivatedChn <-chan localTypes.FOREIGN_ORDER_TYPE,
	orderTimedOutChn chan<- localTypes.FOREIGN_ORDER_TYPE) {

	var orderTimeouts [localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS]time.Time
	var zeroTime = time.Time{}
	pollOrderTimeoutsTicker := time.NewTicker(ORDER_WATCHDOG_POLL_RATE)

	for {
		select {
		case order := <-orderActivatedChn:
			timeout := orderTimeouts[order.Foreign_order.Floor][order.Foreign_order.Button]
			if timeout.IsZero() {
				orderTimeouts[order.Foreign_order.Floor][order.Foreign_order.Button] = time.Now().Add(localTypes.MAX_TIME_TO_FINISH_ORDER)
			}

		case order := <-orderDeactivatedChn:
			orderTimeouts[order.Foreign_order.Floor][order.Foreign_order.Button] = zeroTime

		case <-pollOrderTimeoutsTicker.C:
			for floor := 0; floor < localTypes.NUM_FLOORS; floor++ {
				for button := 0; button < localTypes.NUM_BUTTONS; button++ {
					timeout := orderTimeouts[floor][button]

					if !timeout.IsZero() && timeout.Before(time.Now()) {
						orderTimeouts[floor][button] = zeroTime
						var order localTypes.ORDER
						order.Floor = floor
						order.Button = localTypes.BUTTON_TYPE(button)

						orderTimedOutChn <- order
					}
				}
			}
		}
	}
}

func CombineHRAInput(
	RxElevInfoChan <-chan localTypes.LOCAL_ELEVATOR_INFO,
	RxNewHallRequestChan <-chan localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan <-chan localTypes.BUTTON_INFO,
	TxHRAInputChan chan<- localTypes.HRAInput) {

	currentHRAInput := newAllFalseHRAInput()

	for {
		select {
		case newElevInfo := <-RxElevInfoChan:
			currentHRAInput.States[newElevInfo.ElevatorID] = newElevInfo
			TxHRAInputChan <- currentHRAInput

		case newHRequest := <-RxNewHallRequestChan:
			if currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button] == 0 {
				currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button] = 1
				TxHRAInputChan <- currentHRAInput
			}

		case finishedHOrder := <-RxFinishedHallOrderChan:
			currentHRAInput.HallRequests[finishedHOrder.Floor][finishedHOrder.Button] = 0
			TxHRAInputChan <- currentHRAInput

		default:

		}
	}
}

func newAllFalseHRAInput() localTypes.HRAInput {
	output := localTypes.HRAInput{
		HallRequests: make([localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool),
	}

	for i := range output.HallRequests {
		for j := range output.HallRequests[i] {
			output.HallRequests[i][j] = false
		}
	}
	return output
}

func ReassignOrders(newHRAInput localTypes.HRAInput, hraExecutable string) map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool {
	jsonBytes, err := json.Marshal(newHRAInput)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return
	}

	ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}

	output := map[string][localTypes.NUM_FLOORS][2]bool{}
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return
	}
	return output
}
