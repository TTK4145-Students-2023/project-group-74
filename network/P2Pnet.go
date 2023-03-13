package network

import (
	"fmt"
	"os"
	"project-group-74/network/subs/bcast"
	"project-group-74/network/subs/localip"
	"project-group-74/network/subs/peers"
)

//************ const for P2P ************

const (
	PeerPort  = 15647
	StatePort = 16569
)

// ************** main P2P func *************
func P2Pnet() {
	var id string
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprint("peer-%s-%d", localIP, os.Getpid())
	}

	peerUpdateCh := make(chan peers.PeerUpdate) //We make a channel for receiving updates on the id's of the peers that are alive on the network
	peerTxEnable := make(chan bool)             //Channel to enable

	go peers.Transmitter(PeerPort, id, peerTxEnable)
	go peers.Receiver(PeerPort, peerUpdateCh)

	LocalElevatorInfoTx := make(chan types.LOCAL_ELEVATOR_INFO)
	ForeignElevatorInfoRx := make(chan types.FOREIGN_ELEVATOR_INFO)

	go bcast.Transmitter(StatePort, LocalElevatorInfoTx)
	go bcast.Receiver(StatePort, ForeignElevatorInfoRx)

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:+n")
			fmt.Printf(" Peers:  %q\n", p.Peers)
			fmt.Printf(" New: %q\n", p.New)
			fmt.Printf(" Lost: %q\n", p.Lost)

		case BroadcastLocalState := <-LocalElevatorInfoTx:
			LocalElevatorInfoTx <- BroadcastLocalState

		case ReceiveForeignElevatorState := <-ForeignElevatorInfoRx:
			ForeignElevatorInfoRx <- ReceiveForeignElevatorState
		}
	}
}
