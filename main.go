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

	TxNewHallRequestChan := make(chan localTypes.BUTTON_INFO)
	RxNewHallRequestChan := make(chan localTypes.BUTTON_INFO)

	TxFinishedHallOrderChan := make(chan localTypes.BUTTON_INFO, 10)
	RxFinishedHallOrderChan := make(chan localTypes.BUTTON_INFO, 10)

	TxNewOrdersChan := make(chan map[string]localTypes.HMATRIX, 10)
	RxNewOrdersChan := make(chan map[string]localTypes.HMATRIX, 10)

	TxP2PElevInfoChan := make(chan localTypes.P2P_ELEV_INFO, 10)
	RxP2PElevInfoChan := make(chan localTypes.P2P_ELEV_INFO, 10)

	NewBtnPressChan := make(chan localTypes.BUTTON_INFO, 10)
	NewFloorChan := make(chan int, 10)
	ObstructionChan := make(chan bool, 10)

	TxHRAInputChan := make(chan localTypes.HRAInput, 10)
	RxHRAInputChan := make(chan localTypes.HRAInput, 10)

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
		RxP2PElevInfoChan,
		TxHRAInputChan,
		RxHRAInputChan,
	)

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
		NewBtnPressChan,
		ObstructionChan,
	)

	go decision.OrderAssigner(
		RxElevInfoChan,
		RxNewHallRequestChan,
		RxFinishedHallOrderChan,
		TxNewOrdersChan,
		RxNewOrdersChan,
		TxHRAInputChan,
		RxHRAInputChan)

	go elevio.PollButtons(NewBtnPressChan)
	go elevio.PollFloorSensor(NewFloorChan)
	go elevio.PollObstructionSwitch(ObstructionChan)

	select {}

}
