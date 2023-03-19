package main

import (
	"project-group-74\network"
    "project-group-74\decision"
    "project-group-74\elev_control"
)

func main() {
    
    TxElevInfoChan := make(chan LOCAL_ELEVATOR_INFO)
    RxElevInfoChan := make(chan LOCAL_ELEVATOR_INFO)
    
    TxNewHallRequestChan := make(chan BUTTON_INFO)
    RxNewHallRequestChan := make(chan BUTTON_INFO)
    
    TxFinishedHallOrderChan := make(chan BUTTON_TYPE)
    RxFinishedHallOrderChan := make(chan BUTTON_TYPE)
    
    TxNewOrdersChan := make(chan map[string][types.NUM_FLOORS][types.NUM_BUTTONS-1]bool)
    RxNewOrdersChan := make(chan map[string][types.NUM_FLOORS][types.NUM_BUTTONS-1]bool)

    TxP2PElevInfoChan := make(chan P2P_ELEV_INFO)
    RxP2PElevInfoChan := make(chan P2P_ELEV_INFO)


    NewBtnPressChan := make(chan BUTTON_INFO)
    NewFloorChan := make(chan int)
    

    //********** Set Master/Slave flags ************//
    myIP := flag.string("My IP", "", "The first IP address")
    flag.Parse()
    master := CompareIPAddr
    

    
    //go network.Network_run(NetworkChns)

    go RunElevator(
        TxElevInfoChan,RxElevInfoChan,TxNewHallRequestChan,RxNewHallRequestChan,
        TxFinishedHallOrderChan,RxFinishedHallOrderChan,RxNewOrdersChan,
        TxP2PElevInfoChan,RxP2PElevInfoChan)

    go OrderAssigner(
        RxElevInfoChan,RxNewHallRequestChan,RxFinishedHallOrderChan,
        TxNewOrdersChan,RxNewOrdersChan)

    //go network.Network_run(NetworkChns)

    go elevio.PollButtons(ElevControlChns.NewBtnpress)
    go elevio.PollNewFloor(ElevControlChns.NewFloor)

    //go network.Network(TxP2PElevInfoChan, RxP2PElevInfoChan)
    select{}

}