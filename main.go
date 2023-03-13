package main

import (
	network "project-group-74\network"
    dec "project-group-74\decision"
    elev_control "project-group-74\elev_control"
)

func main() {

    ElevInfoChan := make(chan LOCAL_ELEVATOR_INFO)
    TxP2PElevInfoChan := make(chan P2P_ELEV_INFO)
    RxP2PElevInfoChan := make(chan P2P_ELEV_INFO)
    NewHallRequestChan := make(chan BUTTON_INFO)
    FinishedHallOrderChan := make(chan BUTTON_TYPE)
    NewOrdersChan := make(chan map[string][types.NUM_FLOORS][2]bool)
    

    elev_init()
        decide_master()
  
    go elev_control.Elev_run(ElevControlChns)
    go network.Network_run(NetworkChns)
    go elevio.PollButtons(ElevControlChns.NewBtnpress)
    go elevio.PollNewFloor(ElevControlChns.NewFloor)

    go network.Network(TxP2PElevInfoChan, RxP2PElevInfoChan)


}