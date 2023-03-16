package main

import (
	"project-group-74\network"
    "project-group-74\decision"
    "project-group-74\elev_control"
)

func main() {
    
    ElevInfoChan := make(chan LOCAL_ELEVATOR_INFO)
    NewHallRequestChan := make(chan BUTTON_INFO)
    FinishedHallOrderChan := make(chan BUTTON_TYPE)
    
    NewOrdersChan := make(chan map[string][types.NUM_FLOORS][2]bool)
    NewBtnPressChan := make(chan BUTTON_INFO)
    NewFloorChan := make(chan int)
    
    TxP2PElevInfoChan := make(chan P2P_ELEV_INFO)
    RxP2PElevInfoChan := make(chan P2P_ELEV_INFO)
    
 
    //********** Set Master/Slave flags ************//
    myIP := flag.string("My IP", "", "The first IP address")
    flag.Parse()
    master := CompareIPAddr
    

    elev_init()
  
    go elev_control.Elev_run(ElevinfoChan,TxP2PElevInfoChan,RxP2PElevInfoChan,NewHallRequestChan,FinishedHallOrderChanNewOrdersChan)
    go network.Network_run(NetworkChns)
    go elevio.PollButtons(ElevControlChns.NewBtnpress)
    go elevio.PollNewFloor(ElevControlChns.NewFloor)

    go network.Network(TxP2PElevInfoChan, RxP2PElevInfoChan)


}