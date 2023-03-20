package main

import (
	"elevio"
	"flag"
)

func main() {

	TxElevInfoChan := make(chan LOCAL_ELEVATOR_INFO)
	RxElevInfoChan := make(chan LOCAL_ELEVATOR_INFO)

	TxNewHallRequestChan := make(chan BUTTON_INFO)
	RxNewHallRequestChan := make(chan BUTTON_INFO)

	TxFinishedHallOrderChan := make(chan BUTTON_TYPE)
	RxFinishedHallOrderChan := make(chan BUTTON_TYPE)

	TxNewOrdersChan := make(chan map[string][types.NUM_FLOORS][types.NUM_BUTTONS - 1]bool)
	RxNewOrdersChan := make(chan map[string][types.NUM_FLOORS][types.NUM_BUTTONS - 1]bool)

	TxP2PElevInfoChan := make(chan P2P_ELEV_INFO)
	RxP2PElevInfoChan := make(chan P2P_ELEV_INFO)

	NewBtnPressChan := make(chan BUTTON_INFO)
	NewFloorChan := make(chan int)

	go P2Pnet(
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

	go RunElevator(
		TxElevInfoChan, RxElevInfoChan, TxNewHallRequestChan, RxNewHallRequestChan,
		TxFinishedHallOrderChan, RxFinishedHallOrderChan, RxNewOrdersChan,
		TxP2PElevInfoChan, RxP2PElevInfoChan)

	go OrderAssigner(
		RxElevInfoChan, RxNewHallRequestChan, RxFinishedHallOrderChan,
		TxNewOrdersChan, RxNewOrdersChan)



	go elevio.PollButtons(ElevControlChns.NewBtnpress)
	go elevio.PollNewFloor(ElevControlChns.NewFloor)

	select {}

}
