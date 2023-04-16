package decision_io

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"project-group-74/localTypes"
)

func NewAllFalseHRAInput() localTypes.HRAInput {
	output := localTypes.HRAInput{}
	for i := range output.HallRequests {
		for j := range output.HallRequests[i] {
			output.HallRequests[i][j] = false
		}
	}
	output.States = make(map[string]localTypes.HRAElevState)
	return output
}

func ReassignOrders(newHRAInput localTypes.HRAInput, hraExecutable string) map[string]localTypes.HMATRIX {
	jsonBytes, err := json.Marshal(newHRAInput)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
	}

	ret, err := exec.Command("decision/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
	}

	output := map[string]localTypes.HMATRIX{}
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
	}
	return output
}

func LocalState2HRASTATE(newElevInfo localTypes.LOCAL_ELEVATOR_INFO) localTypes.HRAElevState {
	output := localTypes.HRAElevState{
		State:       elevStateStrings[newElevInfo.State],
		Floor:       newElevInfo.Floor,
		Direction:   motorDirStrings[newElevInfo.Direction],
		CabRequests: newElevInfo.CabCalls,
	}
	return output
}

/*
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
*/