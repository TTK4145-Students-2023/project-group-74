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
) {

	TxHRAInputChan := make(chan DLOCC.HRAInput)

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}
	fmt.Printf(" ORDERASSIGNER RUNNING ")


	go DLOCC.CombineHRAInput(
		RxElevInfoChan,
		RxNewHallRequestChan,
		RxFinishedHallOrderChan,
		TxHRAInputChan)

	for {
		newHRAInput := <-TxHRAInputChan
		fmt.Printf(" ORDERASSIGNER RUNNING ")

		fmt.Printf("")
		if localTypes.IsMaster(localTypes.MyIP, localTypes.PeerList.Peers) {
			newOrders := DLOCC.ReassignOrders(newHRAInput, hraExecutable)

			TxNewOrdersChan <- newOrders
			RxNewOrdersChan <- newOrders
		}
	}
}
