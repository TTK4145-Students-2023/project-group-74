package network

import (
	"fmt"
	"os"
	"project-group-74/network/subs/bcast"
	"project-group-74/network/subs/localip"
	"project-group-74/network/subs/peers"
	"project-group-74/types"
)

//************ const for P2P ************

const (
	PeerPort  = 15647
	StatePort = 16569
)

// ************** main P2P func *************
func P2Pnet() {
	LocalElevatorInfoTx <-chan types.FOREIGN_ELEVETAOR_INFO
	ForeignElevatorInfoRx chan<- types.FOREIGN_ELEVETAOR_INFO


	var MyID string
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		MyID = localIP
		types.ElevInfo[ID] = MyID
		MyID = fmt.Sprint("peer-%s-%d", localIP, os.Getpid())
	}


	
	peerUpdateCh := make(chan peers.PeerUpdate) //We make a channel for receiving updates on the id's of the peers that are alive on the network
	peerTxEnable := make(chan bool)             //Channel to enable

	go peers.Transmitter(PeerPort, id, peerTxEnable)
	go peers.Receiver(PeerPort, peerUpdateCh)

	BCStateTx := make(chan types.FOREIGN_ELEVATOR_INFO)
	StateRx := make(chan types.FOREIGN_ELEVATOR_INFO)

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


func CompareIDAddr (string MyID)bool{
	for range peers.Peers{
		MyID < 
	}
}

func MasterSlave (string MyID){
	if CompareIDAddr(MyID) == true{
		return Master = 1
		else
		return Master = 0
	}
}