package decision

import (
	"fmt"
	"project-group-74/decision/DLOCC"
	"project-group-74/localTypes"
	"runtime"
)

func OrderAssigner(
	RxElevInfoChan <-chan localTypes.LOCAL_ELEVATOR_INFO,
	RxNewHallRequestChan <-chan localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan <-chan localTypes.BUTTON_INFO,
	TxNewOrdersChan chan<- map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	RxNewOrdersChan chan<- map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	TxHRAInputChan <-chan localTypes.HRAInput,
) {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	for {
		newHRAInput, ok := <-TxHRAInputChan
		fmt.Printf("orderAssigner: newHRAInput: ok: %v\n", ok)

		fmt.Printf("")
		if localTypes.IsMaster(localTypes.MyIP, localTypes.PeerList.Peers) {
			newOrders := DLOCC.ReassignOrders(newHRAInput, hraExecutable)
			for k, v := range newOrders {
				fmt.Printf("New Orders: %s: %v\n", k, v)
				if value, ok := newOrders[k]; ok {
					fmt.Printf("This value was passed on TxNewOrdersChan: %v\n", value)
				}
			}
			//RxNewOrdersChan <- newOrders
			TxNewOrdersChan <- newOrders

		}
	}
}
