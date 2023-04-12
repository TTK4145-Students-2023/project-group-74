package main

import (
	"project-group-74/decision"
	"project-group-74/elev_control"
	"project-group-74/elev_control/elevio"
	"project-group-74/localTypes"
	"project-group-74/network"
	"project-group-74/network/subs/localip"
)

func main() {

	TxElevInfoChan := make(chan localTypes.LOCAL_ELEVATOR_INFO, 10)
	RxElevInfoChan := make(chan localTypes.LOCAL_ELEVATOR_INFO, 10)

	TxNewHallRequestChan := make(chan localTypes.BUTTON_INFO, 10)
	RxNewHallRequestChan := make(chan localTypes.BUTTON_INFO, 10)

	TxFinishedHallOrderChan := make(chan localTypes.BUTTON_INFO, 10)
	RxFinishedHallOrderChan := make(chan localTypes.BUTTON_INFO, 10)

	TxNewOrdersChan := make(chan map[string]localTypes.HMATRIX, 10)
	RxNewOrdersChan := make(chan map[string]localTypes.HMATRIX, 10)

	TxP2PElevInfoChan := make(chan localTypes.P2P_ELEV_INFO, 10)
	RxP2PElevInfoChan := make(chan localTypes.P2P_ELEV_INFO, 10)

	NewBtnPressChan := make(chan localTypes.BUTTON_INFO, 10)
	NewFloorChan := make(chan int, 10)

	TxHRAInputChan := make(chan localTypes.HRAInput, 10)

	elevio.Init("localhost:15657", localTypes.NUM_FLOORS)

	myIP, _ := localip.LocalIP()

	go network.P2Pnet(
		TxElevInfoChan,
		RxElevInfoChan,
		TxNewHallRequestChan,
		RxNewHallRequestChan,
		TxFinishedHallOrderChan,
		RxFinishedHallOrderChan,
		TxNewOrdersChan,
		RxNewOrdersChan,
		TxP2PElevInfoChan,
		RxP2PElevInfoChan)

	go elev_control.RunElevator(myIP,
		TxElevInfoChan,
		RxElevInfoChan,
		TxNewHallRequestChan,
		RxNewHallRequestChan,
		TxFinishedHallOrderChan,
		RxFinishedHallOrderChan,
		RxNewOrdersChan,
		TxP2PElevInfoChan,
		RxP2PElevInfoChan,
		NewFloorChan,
		NewBtnPressChan)

	go decision.OrderAssigner(
		RxElevInfoChan,
		RxNewHallRequestChan,
		RxFinishedHallOrderChan,
		TxNewOrdersChan,
		RxNewOrdersChan,
		TxHRAInputChan)

	go elevio.PollButtons(NewBtnPressChan)
	go elevio.PollFloorSensor(NewFloorChan)

	select {}

}
