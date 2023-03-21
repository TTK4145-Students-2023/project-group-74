package network

import (
	"fmt"
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
	TxElevInfoChan 			<-chan 		localTypes.LOCAL_ELEVATOR_INFO,
	RxElevInfoChan 			  chan<- 	localTypes.LOCAL_ELEVATOR_INFO,
	TxNewHallRequestChan 	<-chan 		localTypes.BUTTON_INFO,
	RxNewHallRequestChan 	  chan<- 	localTypes.BUTTON_INFO,
	TxFinishedHallOrderChan <-chan 		localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan   chan<- 	localTypes.BUTTON_INFO,
	TxNewOrdersChan 		<-chan 		map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	RxNewOrdersChan 		  chan<- 	map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	TxP2PElevInfoChan	    <-chan 		localTypes.P2P_ELEV_INFO,
	RxP2PElevInfoChan 		  chan<- 	localTypes.P2P_ELEV_INFO,) {


	if localTypes.MyIP == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "No IP available"
		}
		localTypes.MyIP = localIP
	}

	peerUpdateCh := make(chan peers.PeerUpdate) //We make a channel for receiving updates on the id's of the peers that are alive on the network
	peerTxEnable := make(chan bool)             //Channel to enable

	go peers.Transmitter(PeerPort, localTypes.MyIP, peerTxEnable)
	go peers.Receiver(PeerPort, peerUpdateCh)

	BCLocalStateTx := make(chan localTypes.LOCAL_ELEVATOR_INFO)
	RecieveLocalStateRx := make(chan localTypes.LOCAL_ELEVATOR_INFO)

	go bcast.Transmitter(StatePort, BCLocalStateTx)
	go bcast.Receiver(StatePort, RecieveLocalStateRx)

	BCNewHallReqTx := make(chan localTypes.BUTTON_INFO)
	RecieveNewHallReqRx := make(chan localTypes.BUTTON_INFO)

	go bcast.Transmitter(StatePort, BCNewHallReqTx)
	go bcast.Receiver(StatePort, RecieveNewHallReqRx)

	BCFinHallOrderTx := make(chan localTypes.BUTTON_INFO)
	RecieveFinHallOrderRx := make(chan localTypes.BUTTON_INFO)

	go bcast.Transmitter(StatePort, BCFinHallOrderTx)
	go bcast.Receiver(StatePort, RecieveFinHallOrderRx)

	BCNewOrderTx := make(chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool)
	RecieveOrderRx := make(chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool)

	go bcast.Transmitter(StatePort, BCNewOrderTx)
	go bcast.Receiver(StatePort, RecieveOrderRx)

	BCP2PElevInfo := make(chan localTypes.P2P_ELEV_INFO)
	RecieveP2PElevInfo := make(chan localTypes.P2P_ELEV_INFO)

	go bcast.Transmitter(StatePort, BCP2PElevInfo)
	go bcast.Receiver(StatePort, RecieveP2PElevInfo)

	for {
		select {
	// Print Peer Updates
		case p := <-peerUpdateCh:
			printPeerUpdate(p)
	
	// Broadcasting on network
		case BroadcastLocalState := <-TxElevInfoChan:
			// if !BroadcastLocalState.IsValid(){
			// 	panic("NET: Local elevator info not valid")
			// }
			BCLocalStateTx <- BroadcastLocalState
		case BroadcastNewHallRequest := <- TxNewHallRequestChan:
			BCNewHallReqTx <- BroadcastNewHallRequest
		case BroadcastFinishedHallOrder := <- TxFinishedHallOrderChan:
			BCFinHallOrderTx <- BroadcastFinishedHallOrder
		case BroadcastNewOrders := <- TxNewOrdersChan:
			BCNewOrderTx <- BroadcastNewOrders
		case BroadcastP2PElevInfo := <- TxP2PElevInfoChan:
			BCP2PElevInfo <- BroadcastP2PElevInfo

	// Reading from network
		case ReceiveForeignElevatorState := <- RecieveLocalStateRx:
			// if !ReceiveForeignElevatorState.IsValid(){
			// 	panic("NET: Received data not valid + ??ID??")
			// }
			RxElevInfoChan <- ReceiveForeignElevatorState
		case RecieveNewHallReq := <- RecieveNewHallReqRx:
			RxNewHallRequestChan <- RecieveNewHallReq
		case RecieveFinishedHallOrder := <- RecieveFinHallOrderRx:
			RxFinishedHallOrderChan <- RecieveFinishedHallOrder
		case RecieveNewOrders := <- RecieveOrderRx:
			RxNewOrdersChan <- RecieveNewOrders
		case RecieveP2PElev := <- RecieveP2PElevInfo:
			RxP2PElevInfoChan <- RecieveP2PElev
		}
	}
}

func printPeerUpdate(p peers.PeerUpdate){
	fmt.Printf("Peer update:+n")
	fmt.Printf(" Peers:  %q\n", p.Peers)
	fmt.Printf(" New: %q\n", p.New)
	fmt.Printf(" Lost: %q\n", p.Lost)
}
