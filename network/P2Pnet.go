package network

import (
	"fmt"
	"net"
	"os"
	"flag"
	"project-group-74/network/subs/bcast"
	"project-group-74/network/subs/localip"
	"project-group-74/network/subs/peers"
	"project-group-74/localTypes"
)

//************ const for P2P ************

const (
	PeerPort  = 15647
	StatePort = 16569
)

// ************** main P2P func *************
func P2Pnet(
	TxElevInfoChan <-chan localTypes.LOCAL_ELEVATOR_INFO,
	RxElevInfoChan chan<- localTypes.LOCAL_ELEVATOR_INFO,
	TxNewHallRequestChan <-chan localTypes.BUTTON_INFO,
	RxNewHallRequestChan chan<- localTypes.BUTTON_INFO,
	TxFinishedHallOrderChan <-chan localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan chan<- localTypes.BUTTON_INFO,
	TxNewOrdersChan <-chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	RxNewOrdersChan <-chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	TxP2PElevInfoChan <-chan localTypes.P2P_ELEV_INFO,
	RxP2PElevInfoChan <-chan localTypes.P2P_ELEV_INFO,) {



	var MyIP string
	if MyIP == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		MyIP = localIP
	}


	
	peerUpdateCh := make(chan peers.PeerUpdate) //We make a channel for receiving updates on the id's of the peers that are alive on the network
	peerTxEnable := make(chan bool)             //Channel to enable

	go peers.Transmitter(PeerPort, MyIP, peerTxEnable)
	go peers.Receiver(PeerPort, peerUpdateCh)

	BCStateTx := make(chan localTypes.LOCAL_ELEVATOR_INFO)
	StateRx := make(chan localTypes.LOCAL_ELEVATOR_INFO)

	go bcast.Transmitter(StatePort, BCStateTx)
	go bcast.Receiver(StatePort, StateRx)

	for {
		select {
	// Peer Updates
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:+n")
			fmt.Printf(" Peers:  %q\n", p.Peers)
			fmt.Printf(" New: %q\n", p.New)
			fmt.Printf(" Lost: %q\n", p.Lost)
	
	// Broadcasting on network
		case BroadcastLocalState := <-LocalElevatorInfoTx:
			if !BroadcastLocalState.IsValid(){
				panic("NET: Local elevator info not valid")
			}
			BCStateTx <- BroadcastLocalState

	// Reading from network
		case ReceiveForeignElevatorState := <- StateRx:
			if !ReceiveForeignElevatorState.IsValid(){
				panic("NET: Received data not valid + ??ID??")
			}
			ForeignElevatorInfoRx <- ReceiveForeignElevatorState
		}
	}
}