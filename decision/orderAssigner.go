package decision

import (
	"fmt"
	"project-group-74/decision/DLOCC"
	"project-group-74/localTypes"
	"runtime"
	"time"
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

	currentHRAInput := DLOCC.NewAllFalseHRAInput()

	for {
		select {
		case newElevInfo := <-RxElevInfoChan:
			//if newElevInfo.State !isValid() || !isValidID(newElevInfo.ElevID){
			//	panic("Corrupt elevator data from RxElevInfoChan")
			//}
			newHRAelev := DLOCC.LocalState2HRASTATE(newElevInfo)
			currentHRAInput.States[newElevInfo.ElevID] = newHRAelev

			if localTypes.IsMaster(localTypes.MyIP, localTypes.PeerList.Peers) {
				newOrders := DLOCC.ReassignOrders(currentHRAInput, hraExecutable)
				/*for k, v := range newOrders {
					fmt.Printf("New Orders: %s: %v\n", k, v)
				}*/
				//RxNewOrdersChan <- newOrders
				TxNewOrdersChan <- newOrders
			}

		case newHRequest := <-RxNewHallRequestChan:
			//if !isValidFloor(newHRequest.Floor) || newHRequest.Button !isValid(){
			//	panic("Corrupt elevator data from RxNewHallRequestChan")
			//}
			if !currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button] {
				currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button] = true
				fmt.Printf("DLOCC: NewHrequest: \n")

				if localTypes.IsMaster(localTypes.MyIP, localTypes.PeerList.Peers) {
					newOrders := DLOCC.ReassignOrders(currentHRAInput, hraExecutable)
					for k, v := range newOrders {
						fmt.Printf("New Orders: %s: %v\n", k, v)
					}
					//RxNewOrdersChan <- newOrders
					TxNewOrdersChan <- newOrders

				}

			}

		case finishedHOrder := <-RxFinishedHallOrderChan:
			//if !isValidFloor(finishedHOrder.Floor) || finishedHOrder.Button !isValid(){
			//	panic("Corrupt elevator data from RxFinishedHallOrderChan")
			//}
			fmt.Printf("DLOCC: new finishedHOrder: %v\n \n", currentHRAInput)
			currentHRAInput.HallRequests[finishedHOrder.Floor][finishedHOrder.Button] = false

			if localTypes.IsMaster(localTypes.MyIP, localTypes.PeerList.Peers) {
				newOrders := DLOCC.ReassignOrders(currentHRAInput, hraExecutable)
				for k, v := range newOrders {
					fmt.Printf("New Orders: %s: %v\n", k, v)
				}
				//RxNewOrdersChan <- newOrders
				TxNewOrdersChan <- newOrders

			}

		default:
			time.Sleep((time.Millisecond * 200))
		}

	}
}
