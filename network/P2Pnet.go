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
	recieveTimer := time.NewTimer(1)
	recieveTimer.Stop()

	for {
		select {
		// Print Peer Updates
		case p := <-peerUpdateCh:
			printPeerUpdate(p)
			localTypes.PeerList.Peers = p.Peers
			fmt.Printf("This is PeerList: %q\n", localTypes.PeerList.Peers)

			// Broadcasting on network
		case <-broadcastTimer.C:
			bc, ok := <-TxElevInfoChan
			if ok {
				fmt.Printf("NET.BC: Local State\n")
				BCLocalStateTx <- bc
			}
			bc1, ok2 := <-TxNewHallRequestChan
			if ok2 {
				fmt.Printf("NET.BC: New Hall req\n")
				BCNewHallReqTx <- bc1
			}
			bc2, ok3 := <-TxFinishedHallOrderChan
			if ok3 {
				fmt.Printf("NET.BC: Finished hall order\n")
				BCFinHallOrderTx <- bc2
			}
			bc3, ok4 := <-TxNewOrdersChan
			if ok4 {
				fmt.Printf("NET.BC: new order\n")
				BCNewOrderTx <- bc3
			}
			bc4, ok5 := <-TxP2PElevInfoChan
			if ok5 {
				fmt.Printf("NET.BC: Elev Info\n")
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
		case <-recieveTimer.C:
			r, ok6 := <-RecieveLocalStateRx
			if ok6 {
				fmt.Printf("NET Recived elevinfo\n")
				RxElevInfoChan <- r

			}
			r1, ok7 := <-RecieveNewHallReqRx
			if ok7 {
				fmt.Printf("NET Recived new hall req\n")
				RxNewHallRequestChan <- r1
			}
			r2, ok8 := <-RecieveFinHallOrderRx
			if ok8 {
				fmt.Printf("NET Recived finished hall order\n")
				RxFinishedHallOrderChan <- r2
			}
			r3, ok9 := <-RecieveOrderRx
			if ok9 {
				fmt.Printf("NET Recived new orders\n")
				RxNewOrdersChan <- r3
			}
			r4, ok10 := <-RecieveP2PElevInfo
			if ok10 {
				fmt.Printf("NET Recived P2Pelevinfo\n")
				RxP2PElevInfoChan <- r4
			}
			recieveTimer.Reset(1)

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
