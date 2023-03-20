package main

import (
	"project-group-74/localTypes"
	"project-group-74/network"
	"project-group-74/decision"
	"project-group-74/elev_control"
	"project-group-74/elev_control/elevio"



)

func main() {

	TxElevInfoChan := make(chan localTypes.LOCAL_ELEVATOR_INFO)
	RxElevInfoChan := make(chan localTypes.LOCAL_ELEVATOR_INFO)

	TxNewHallRequestChan := make(chan localTypes.BUTTON_INFO)
	RxNewHallRequestChan := make(chan localTypes.BUTTON_INFO)

	TxFinishedHallOrderChan := make(chan localTypes.BUTTON_INFO)
	RxFinishedHallOrderChan := make(chan localTypes.BUTTON_INFO)

	TxNewOrdersChan := make(chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool)
	RxNewOrdersChan := make(chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool)

	TxP2PElevInfoChan := make(chan localTypes.P2P_ELEV_INFO)
	RxP2PElevInfoChan := make(chan localTypes.P2P_ELEV_INFO)

	NewBtnPressChan := make(chan localTypes.BUTTON_INFO)
	NewFloorChan := make(chan int)

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

	go elev_control.RunElevator(
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
		RxNewOrdersChan)



	go elevio.PollButtons(NewBtnPressChan)
	go elevio.PollFloorSensor(NewFloorChan)

	select {}

}
