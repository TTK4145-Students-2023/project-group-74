package network

import (
	"fmt"
	"project-group-74/localTypes"
	"project-group-74/network/subs/bcast"
	"project-group-74/network/subs/localip"
	"project-group-74/network/subs/peers"
	"time"
)

// ************ const for P2P ************
const (
	PeerPort      = 15699
	StatePort     = 16599
	BroadcastRate = 2 * time.Second
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
	RxNewOrdersChan chan<- map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	TxP2PElevInfoChan <-chan localTypes.P2P_ELEV_INFO,
	RxP2PElevInfoChan chan<- localTypes.P2P_ELEV_INFO) {

	if localTypes.MyIP == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "No IP available"
		}
		localTypes.MyIP = localIP
	}
	fmt.Printf(" NETWORK RUNNING\n")

	peerUpdateCh := make(chan peers.PeerUpdate) //We make a channel for receiving updates on the id's of the peers that are alive on the network
	peerTxEnable := make(chan bool)             //Channel to enable

	go peers.Transmitter(PeerPort, localTypes.MyIP, peerTxEnable)
	go peers.Receiver(PeerPort, peerUpdateCh)

	var (
		BCLocalStateTx        = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		RecieveLocalStateRx   = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		BCNewHallReqTx        = make(chan localTypes.BUTTON_INFO)
		RecieveNewHallReqRx   = make(chan localTypes.BUTTON_INFO)
		BCFinHallOrderTx      = make(chan localTypes.BUTTON_INFO)
		RecieveFinHallOrderRx = make(chan localTypes.BUTTON_INFO)
		BCNewOrderTx          = make(chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool)
		RecieveOrderRx        = make(chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool)
		BCP2PElevInfo         = make(chan localTypes.P2P_ELEV_INFO)
		RecieveP2PElevInfo    = make(chan localTypes.P2P_ELEV_INFO)
	)
	// GoRoutines to recieve from NTW
	go bcast.Receiver(StatePort, RecieveLocalStateRx)
	go bcast.Receiver(StatePort, RecieveNewHallReqRx)
	go bcast.Receiver(StatePort, RecieveFinHallOrderRx)
	go bcast.Receiver(StatePort, RecieveOrderRx)
	go bcast.Receiver(StatePort, RecieveP2PElevInfo)

	// GoRoutines to broadcast over NTW
	go bcast.Transmitter(StatePort, BCLocalStateTx)
	go bcast.Transmitter(StatePort, BCNewHallReqTx)
	go bcast.Transmitter(StatePort, BCFinHallOrderTx)
	go bcast.Transmitter(StatePort, BCNewOrderTx)
	go bcast.Transmitter(StatePort, BCP2PElevInfo)

	// Broadcast Timer
	broadcastTimer := time.NewTimer(BroadcastRate)

	for {
		select {
		// Print Peer Updates
		case p := <-peerUpdateCh:
			printPeerUpdate(p)
			localTypes.PeerList.Peers = p.Peers

			// Broadcasting on network
		case <-broadcastTimer.C:
			bc, ok := <-TxElevInfoChan
			if ok{
				BCLocalStateTx <- bc
			}
			bc1, ok := <-TxNewHallRequestChan
			if ok{
				BCNewHallReqTx <- bc1
			}
			bc2, ok := <-TxFinishedHallOrderChan
			if ok{
				BCFinHallOrderTx <- bc2
			}
			bc3, ok := <-TxNewOrdersChan
			if ok{
				BCNewOrderTx <- bc3
			}
			bc4, ok := <-TxP2PElevInfoChan
			if ok{
				BCP2PElevInfo <- bc4
			}
			broadcastTimer.Reset(BroadcastRate)

			// for {
			// 	select {
			// 	case BroadcastElevInfo := <-TxElevInfoChan:
			// 		// if !BroadcastLocalState.IsValid(){
			// 		// 	panic("NET: Local elevator info not valid")
			// 		// }
			// 		BCLocalStateTx <- BroadcastElevInfo
			// 	case BroadcastNewHallRequest := <-TxNewHallRequestChan:
			// 		BCNewHallReqTx <- BroadcastNewHallRequest
			// 	case BroadcastFinishedHallOrder := <-TxFinishedHallOrderChan:
			// 		BCFinHallOrderTx <- BroadcastFinishedHallOrder
			// 	case BroadcastNewOrders := <-TxNewOrdersChan:
			// 		BCNewOrderTx <- BroadcastNewOrders
			// 	case BroadcastP2PElevInfo := <-TxP2PElevInfoChan:
			// 		BCP2PElevInfo <- BroadcastP2PElevInfo
			// 	default:
			// 		fmt.Printf("Reset broadcastTimer")
			// 		broadcastTimer.Reset(BroadcastRate)
			// 	}
			// }

			// Reading from network
		// case ReceiveForeignElevatorState := <-RecieveLocalStateRx:
		// 	// if !ReceiveForeignElevatorState.IsValid(){
		// 	// 	panic("NET: Received data not valid + ??ID??")
		// 	// }
		// 	RxElevInfoChan <- ReceiveForeignElevatorState
		// case RecieveNewHallReq := <-RecieveNewHallReqRx:
		// 	RxNewHallRequestChan <- RecieveNewHallReq
		// case RecieveFinishedHallOrder := <-RecieveFinHallOrderRx:
		// 	RxFinishedHallOrderChan <- RecieveFinishedHallOrder
		// case RecieveNewOrders := <-RecieveOrderRx:
		// 	RxNewOrdersChan <- RecieveNewOrders
		// case RecieveP2PElev := <-RecieveP2PElevInfo:
		// 	RxP2PElevInfoChan <- RecieveP2PElev
		}
	}
}

func printPeerUpdate(p peers.PeerUpdate) {
	fmt.Printf("Peer update:\n")
	fmt.Printf(" Peers:  %q\n", p.Peers)
	fmt.Printf(" New: %q\n", p.New)
	fmt.Printf(" Lost: %q\n", p.Lost)
}
