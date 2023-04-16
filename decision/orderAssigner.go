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
	TxNewOrdersChan chan<- map[string]localTypes.HMATRIX,
	RxNewOrdersChan chan<- map[string]localTypes.HMATRIX,
	TxHRAInputChan chan<- localTypes.HRAInput,
	RxHRAInputChan <-chan localTypes.HRAInput,
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
	var lastOrders map[string]localTypes.HMATRIX
	//lastOrders := DLOCC.ReassignOrders(currentHRAInput, hraExecutable)
	//var OAticker = time.NewTicker(time.Millisecond * 100)
	restored := false
	initializing := true
	var initimer *time.Timer
	initimer = time.NewTimer(time.Second * 1)
	for initializing {
		select {
		case NewHRAInput := <-RxHRAInputChan:
			if !restored{
				currentHRAInput = NewHRAInput
				restored = true
				fmt.Printf("\nNew HRAInput into init \n")
			}

		case <-initimer.C:
			initializing = false
			fmt.Printf("\n\n\n\nInitializing finished!\n\n\n")
		default:
			time.Sleep(80 * time.Millisecond)
		}
	}
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
				//if !reflect.DeepEqual(newOrders, lastOrders) {
				lastOrders = newOrders

				if len(localTypes.PeerList.Peers) == 0 {
					RxNewOrdersChan <- lastOrders
				} else {
					TxNewOrdersChan <- lastOrders
				}
				/*for k, v := range newOrders {
					fmt.Printf("New Orders from newelevinfo: %s: %v\n", k, v)
				}*/
				//}
			}else{
			fmt.Printf("\n pre blocking\n")
			TxHRAInputChan <- currentHRAInput
			fmt.Printf("\n non blocking\n")}

		case newHRequest := <-RxNewHallRequestChan:
			//if !isValidFloor(newHRequest.Floor) || newHRequest.Button !isValid(){
			//	panic("Corrupt elevator data from RxNewHallRequestChan")
			//}
			if !currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button] {
				currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button] = true
				fmt.Printf("DLOCC: NewHrequest: \n")
				if localTypes.IsMaster(localTypes.MyIP, localTypes.PeerList.Peers) {
					newOrders := DLOCC.ReassignOrders(currentHRAInput, hraExecutable)
					//if !reflect.DeepEqual(newOrders, lastOrders) {
					lastOrders = newOrders

					if len(localTypes.PeerList.Peers) == 0 {
						RxNewOrdersChan <- lastOrders
					} else {
						TxNewOrdersChan <- lastOrders
					}
					for k, v := range newOrders {
						fmt.Printf("New from hallreq: %s: %v\n", k, v)
					}
					//}
				}
				}else{
					fmt.Printf("\n pre blocking\n")
					TxHRAInputChan <- currentHRAInput
					fmt.Printf("\n non blocking\n")}

		case finishedHOrder := <-RxFinishedHallOrderChan:
			//if !isValidFloor(finishedHOrder.Floor) || finishedHOrder.Button !isValid(){
			//	panic("Corrupt elevator data from RxFinishedHallOrderChan")
			//}
			fmt.Printf("DLOCC: new finishedHOrder: %v\n \n", currentHRAInput)
			currentHRAInput.HallRequests[finishedHOrder.Floor][finishedHOrder.Button] = false
			if localTypes.IsMaster(localTypes.MyIP, localTypes.PeerList.Peers) {
				newOrders := DLOCC.ReassignOrders(currentHRAInput, hraExecutable)
				//if !reflect.DeepEqual(newOrders, lastOrders) {
				lastOrders = newOrders

				if len(localTypes.PeerList.Peers) == 0 {
					RxNewOrdersChan <- lastOrders
				} else {
					TxNewOrdersChan <- lastOrders
				}
				//for k, v := range newOrders {
				fmt.Printf("New Orders finhallreq: %+v\n", newOrders)
				//}
				//}
				}else{
					fmt.Printf("\n pre blocking\n")
					TxHRAInputChan <- currentHRAInput
					fmt.Printf("\n non blocking\n")}

			/*
				case <-OAticker.C:
					if localTypes.IsMaster(localTypes.MyIP, localTypes.PeerList.Peers) {
						newOrders := DLOCC.ReassignOrders(currentHRAInput, hraExecutable)
						if !reflect.DeepEqual(newOrders, lastOrders) {
							lastOrders = newOrders

							if len(localTypes.PeerList.Peers) == 0 {
								RxNewOrdersChan <- lastOrders
							} else {
								TxNewOrdersChan <- lastOrders
							}
							for k, v := range newOrders {
								fmt.Printf("New Orders from ticker: %s: %v\n", k, v)
							}
						}
					}*/
		default:
			time.Sleep((time.Millisecond * 100))
		}

	}
}
