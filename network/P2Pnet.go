package network

import (
	"flag"
	"os"
	"fmt"
	"time"
	"project-group-74/elev_control/elevio"
	"project-group-74/network/subs/bcast"
	"project-group-74/network/subs/conn"
	"project-group-74/network/subs/localip"
	"project-group-74/network/subs/peers"
)

//************ const for P2P ************

const (
	PeerPort = 15647
	StatePort = 16569
)

//************ Variables for P2P ************

type LocalState struct{
	Floor int
	Dir elev_control.MotorDirection
	State elev_control.ElevState
	Orders [3][4]bool
}


//************** main P2P func *************
func P2Pnet(){
	var id string
	if id == ""{
		localIP, err := localip.LocalIP()
		if err != nil{
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprint("peer-%s-%d", localIP, os.Getpid())
	}

	peerUpdateCh := make(chan peers.PeerUpdate) //We make a channel for receiving updates on the id's of the peers that are alive on the network
	peerTxEnable := make(chan bool) //Channel to enable 

	go peers.Transmitter(PeerPort, id, peerTxEnable)
	go peers.Receiver(PeerPort, peerUpdateCh)



	go bcast.Transmitter(StatePort, )
	go bcast.Receiver(StatePort, )
}