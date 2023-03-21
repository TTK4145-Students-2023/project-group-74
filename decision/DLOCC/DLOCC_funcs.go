package DLOCC

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"project-group-74/localTypes"
	"time"
)

func orderWatchdog(
	orderActivatedChn <-chan localTypes.BUTTON_INFO,
	orderDeactivatedChn <-chan localTypes.BUTTON_INFO,
	orderTimedOutChn chan<- localTypes.BUTTON_INFO) {

	var orderTimeouts [localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS]time.Time
	var zeroTime = time.Time{}
	pollOrderTimeoutsTicker := time.NewTicker(localTypes.ORDER_WATCHDOG_POLL_RATE)

	for {
		select {
		case order := <-orderActivatedChn:
			timeout := orderTimeouts[order.Floor][order.Button]
			if timeout.IsZero() {
				orderTimeouts[order.Floor][order.Button] = time.Now().Add(localTypes.MAX_TIME_TO_FINISH_ORDER)
			}

		case order := <-orderDeactivatedChn:
			orderTimeouts[order.Floor][order.Button] = zeroTime

		case <-pollOrderTimeoutsTicker.C:
			for floor := 0; floor < localTypes.NUM_FLOORS; floor++ {
				for button := 0; button < localTypes.NUM_BUTTONS; button++ {
					timeout := orderTimeouts[floor][button]

					if !timeout.IsZero() && timeout.Before(time.Now()) {
						orderTimeouts[floor][button] = zeroTime
						var order localTypes.BUTTON_INFO
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
	TxHRAInputChan chan<- HRAInput) {

	currentHRAInput := newAllFalseHRAInput()

	for {
		select {
		case newElevInfo := <-RxElevInfoChan:
			//if newElevInfo.State !isValid() || !isValidID(newElevInfo.ElevID){
			//	panic("Corrupt elevator data from RxElevInfoChan")
			//}
			newHRAelev := localState2HRASTATE(newElevInfo)
			currentHRAInput.States[newElevInfo.ElevID] = newHRAelev
			TxHRAInputChan <- currentHRAInput

		case newHRequest := <-RxNewHallRequestChan:
			//if !isValidFloor(newHRequest.Floor) || newHRequest.Button !isValid(){
			//	panic("Corrupt elevator data from RxNewHallRequestChan")
			//}
			if !currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button] {
				currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button] = true
				TxHRAInputChan <- currentHRAInput
			}

		case finishedHOrder := <-RxFinishedHallOrderChan:
			//if !isValidFloor(finishedHOrder.Floor) || finishedHOrder.Button !isValid(){
			//	panic("Corrupt elevator data from RxFinishedHallOrderChan")
			//}
			currentHRAInput.HallRequests[finishedHOrder.Floor][finishedHOrder.Button] = false
			TxHRAInputChan <- currentHRAInput

		default:

		}
	}
}

func newAllFalseHRAInput() HRAInput {
	output := HRAInput{}
	for i := range output.HallRequests {
		for j := range output.HallRequests[i] {
			output.HallRequests[i][j] = false
		}
	}
	output.States = make(map[string]HRAElevState)
	return output
}

func ReassignOrders(newHRAInput HRAInput, hraExecutable string) map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool {
	jsonBytes, err := json.Marshal(newHRAInput)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
	}

	ret, err := exec.Command("project-group-74/decision/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
	}

	output := map[string][localTypes.NUM_FLOORS][2]bool{}
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
	}
	return output
}

func localState2HRASTATE(newElevInfo localTypes.LOCAL_ELEVATOR_INFO) HRAElevState {
	output := HRAElevState{
		State:       getElevStateString(newElevInfo.State),
		Floor:       newElevInfo.Floor,
		Direction:   getMotorDirString(newElevInfo.Direction),
		CabRequests: newElevInfo.CabCalls,
	}
	return output
}

func getMotorDirString(md localTypes.MOTOR_DIR) string {
	return motorDirStrings[md]
}

func getElevStateString(state localTypes.ELEVATOR_STATE) string {
	return elevStateStrings[state]
}
