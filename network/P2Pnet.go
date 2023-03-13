package network

import (
	"fmt"
	"net"
	"os"
	"flag"
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
func P2Pnet(
	LocalElevatorInfoTx <-chan types.FOREIGN_ELEVETAOR_INFO,
	ForeignElevatorInfoRx chan<- types.FOREIGN_ELEVETAOR_INFO) {



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

	go peers.Transmitter(PeerPort, MyID, peerTxEnable)
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


func splitIPAddr (ip string)byte{
	addr := net.ParseIP(ip).To4()
	return addr[3]
}

type ipsFlag []string

func CompareIPAddr (MyID string, Peers []string)bool{
	var ips ipsFlag
	flag.Var(&ips, "ip", "list of IP addresses")

	flag.Parse()

	lowestIP := Peers[0]
	for _, ip := range peers.Peers[1:]{
		lastOctet := splitIPAddr(ip)
		addrLowest := net.ParseIP(lowestIP).To4()
		if lastOctet < addrLowest[3]{
			lowestIP = ip
		}
	}
	myIP := net.ParseIP(MyID).To4()
	lowestIP = string(net.ParseIP(lowestIP).To4())
	return myIP[3] <= lowestIP[3]
}

func setMasterSlaveFlag (MyID, Peers []string){
	for _, peer := range peers.Peers[0:]{
		if CompareIPAddr(MyID, peers.Peers) == true {
			master := MyID
		}else{
			slave := MyID
		}
	}
}