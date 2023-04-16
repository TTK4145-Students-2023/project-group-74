package orderAssigner

import (
	"project-group-74/localTypes"
	"project-group-74/network"
	"project-group-74/orderAssigner/decision_io"
	"runtime"
	"time"
)

// ----- MAIN FUNCTION (ORDER ASSIGNER) ------ //
func OrderAssigner(
	RxElevInfoChan <-chan localTypes.LOCAL_ELEVATOR_INFO,
	RxHallRequestChan <-chan localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan <-chan localTypes.BUTTON_INFO,
	TxNewOrdersChan chan<- map[string]localTypes.HMATRIX,
	RxNewOrdersChan chan<- map[string]localTypes.HMATRIX,
	TxHRAInputChan <-chan localTypes.HRAInput) {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	currentHRAInput := decision_io.NewAllFalseHRAInput()
	for {
		select {
		case ElevInfo := <-RxElevInfoChan:
			if !(ElevInfo.State.IsValid()) || !(localTypes.IsValidID(ElevInfo.ElevID)) {
				panic("Corrupt elevator data from RxElevInfoChan to Order Assigner")
			}
			currentHRAInput.States[ElevInfo.ElevID] = decision_io.LocalState2HRASTATE(ElevInfo)
			if network.IsMaster(network.MyIP, network.PeerList.Peers) {
				newOrders := decision_io.ReassignOrders(currentHRAInput, hraExecutable)
				network.SendNewOrders(newOrders, RxNewOrdersChan, TxNewOrdersChan)
			}

		case HallRequest := <-RxHallRequestChan:
			if !(localTypes.IsValidFloor(HallRequest.Floor)) || !(HallRequest.Button.IsValid()) {
				panic("Corrupt elevator data from RxHallRequestChan to Order Assigner")
			}
			if !(currentHRAInput.HallRequests[HallRequest.Floor][HallRequest.Button]) {
				currentHRAInput.HallRequests[HallRequest.Floor][HallRequest.Button] = true
				if network.IsMaster(network.MyIP, network.PeerList.Peers) {
					newOrders := decision_io.ReassignOrders(currentHRAInput, hraExecutable)
					network.SendNewOrders(newOrders, RxNewOrdersChan, TxNewOrdersChan)
				}
			}
		case finHallOrder := <-RxFinishedHallOrderChan:
			if !(localTypes.IsValidFloor(finHallOrder.Floor)) || !(finHallOrder.Button.IsValid()) {
				panic("Corrupt elevator data from RxFinishedHallOrderChan to Order Assigner")
			}
			currentHRAInput.HallRequests[finHallOrder.Floor][finHallOrder.Button] = false
			if network.IsMaster(network.MyIP, network.PeerList.Peers) {
				newOrders := decision_io.ReassignOrders(currentHRAInput, hraExecutable)
				network.SendNewOrders(newOrders, RxNewOrdersChan, TxNewOrdersChan)
			}

		default:
			time.Sleep((time.Millisecond * 100))
		}

	}
}
