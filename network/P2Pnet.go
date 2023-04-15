package network

import (
	"fmt"
	"project-group-74/localTypes"
	"project-group-74/network/subs/bcast"
	"project-group-74/network/subs/localip"
	"project-group-74/network/subs/peers"
	"reflect"
	"sort"
	"time"
)

// ************ const for P2P ************
const (
	PeerPort      = 15699
	Port1         = 16599
	Port2         = 17000
	port3         = 18000
	port4         = 19000
	port5         = 20000
	BroadcastRate = 100 * time.Millisecond
)

// ************** main P2P func *************
func P2Pnet(
	TxElevInfoChan <-chan localTypes.LOCAL_ELEVATOR_INFO,
	RxElevInfoChan chan<- localTypes.LOCAL_ELEVATOR_INFO,
	TxNewHallRequestChan <-chan localTypes.BUTTON_INFO,
	RxNewHallRequestChan chan<- localTypes.BUTTON_INFO,
	TxFinishedHallOrderChan <-chan localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan chan<- localTypes.BUTTON_INFO,
	TxNewOrdersChan <-chan map[string]localTypes.HMATRIX,
	RxNewOrdersChan chan<- map[string]localTypes.HMATRIX,
	TxP2PElevInfoChan <-chan localTypes.P2P_ELEV_INFO,
	RxP2PElevInfoChan chan<- localTypes.P2P_ELEV_INFO) {

	var (
		// Tx var
		p2pElevInfo = localTypes.P2P_ELEV_INFO{}
		newHallReq  = localTypes.BUTTON_INFO{}
		finHallReq  = localTypes.BUTTON_INFO{}
		localState  = localTypes.LOCAL_ELEVATOR_INFO{}
		newOrder    = map[string]localTypes.HMATRIX{}
		// Rx var
		rxP2pElevinfo = localTypes.P2P_ELEV_INFO{}
		rxnewHallReq  = localTypes.BUTTON_INFO{}
		rxfinHallReq  = localTypes.BUTTON_INFO{}
		rxLocalState  = localTypes.LOCAL_ELEVATOR_INFO{}
		rxnewOrder    = map[string]localTypes.HMATRIX{}
		// Tx chan
		BCLocalStateTx   = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		BCNewHallReqTx   = make(chan localTypes.BUTTON_INFO)
		BCFinHallOrderTx = make(chan localTypes.BUTTON_INFO)
		BCNewOrderTx     = make(chan map[string]localTypes.HMATRIX)
		BCP2PElevInfoTx  = make(chan localTypes.P2P_ELEV_INFO)
		// Rx chan
		RecieveLocalStateRx   = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		RecieveNewHallReqRx   = make(chan localTypes.BUTTON_INFO)
		RecieveFinHallOrderRx = make(chan localTypes.BUTTON_INFO)
		RecieveOrderRx        = make(chan map[string]localTypes.HMATRIX)
		RecieveP2PElevInfo    = make(chan localTypes.P2P_ELEV_INFO)
	)

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

	// GoRoutines to recieve from NTW
	go bcast.Receiver(Port1, RecieveLocalStateRx)
	go bcast.Receiver(Port2, RecieveNewHallReqRx)
	go bcast.Receiver(port3, RecieveFinHallOrderRx)
	go bcast.Receiver(port4, RecieveOrderRx)
	go bcast.Receiver(port5, RecieveP2PElevInfo)

	// GoRoutines to broadcast over NTW
	go bcast.Transmitter(Port1, BCLocalStateTx)
	go bcast.Transmitter(Port2, BCNewHallReqTx)
	go bcast.Transmitter(port3, BCFinHallOrderTx)
	go bcast.Transmitter(port4, BCNewOrderTx)
	go bcast.Transmitter(port5, BCP2PElevInfoTx)

	// Broadcast Timer
	//broadcastTimer := time.NewTimer(BroadcastRate)
	recieveTimer := time.NewTimer(1)
	recieveTimer.Stop()

	for {
		select {

		// Print Peer Updates
		case p := <-peerUpdateCh:
			printPeerUpdate(p)
			localTypes.PeerList.Peers = p.Peers
			// Broadcasting on network
		case localState = <-TxElevInfoChan:
			fmt.Printf("NET: Transmit local state::   %+v\n", localState)
			BCLocalStateTx <- localState
		case newHallReq = <-TxNewHallRequestChan:
			fmt.Printf("NET: Transmit new hall req::   %+v\n", rxnewHallReq)
			BCNewHallReqTx <- newHallReq
		case finHallReq = <-TxFinishedHallOrderChan:
			fmt.Printf("NET: Transmit finished hall order::   %+v\n", finHallReq)
			BCFinHallOrderTx <- finHallReq
		case newOrder = <-TxNewOrdersChan:
			fmt.Printf("NET: Transmit new order::   %+v\n", newOrder)
			BCNewOrderTx <- newOrder
		case p2pElevInfo = <-TxP2PElevInfoChan:
			fmt.Printf("NET: Transmit P2Pelevinfo::   %+v\n", p2pElevInfo)
			BCP2PElevInfoTx <- p2pElevInfo
		/*case <-broadcastTimer.C:
		fmt.Printf("NET: Broadcasting NOW\n")
		BCLocalStateTx <- localState
		BCNewHallReqTx <- newHallReq
		BCFinHallOrderTx <- finHallReq
		BCNewOrderTx <- newOrder
		BCP2PElevInfoTx <- p2pElevInfo
		broadcastTimer.Reset(BroadcastRate)*/

		case newrxP2pElevinfo := <-RecieveP2PElevInfo: // Legge på sender ID?
			sort.Slice(rxP2pElevinfo, func(i, j int) bool {
				return rxP2pElevinfo[i].ElevID < rxP2pElevinfo[j].ElevID
			})
			sort.Slice(newrxP2pElevinfo, func(i, j int) bool {
				return newrxP2pElevinfo[i].ElevID < newrxP2pElevinfo[j].ElevID
			})
			if !reflect.DeepEqual(rxP2pElevinfo, newrxP2pElevinfo) {
				rxP2pElevinfo = newrxP2pElevinfo
				fmt.Printf("NET:P2Pelevinfo:: %+v\n", rxP2pElevinfo)
				RxP2PElevInfoChan <- rxP2pElevinfo
			}
		case newrxnewHallReq := <-RecieveNewHallReqRx:
			if rxnewHallReq != newrxnewHallReq {
				rxnewHallReq = newrxnewHallReq
				fmt.Printf("NET:newHallReq:: %+v\n", rxnewHallReq)
				RxNewHallRequestChan <- rxnewHallReq
			}
		case newrxfinHallReq := <-RecieveFinHallOrderRx:
			if rxfinHallReq != newrxfinHallReq {
				rxfinHallReq = newrxfinHallReq
				fmt.Printf("NET:FinHallReq:: %+v\n", rxfinHallReq)
				RxFinishedHallOrderChan <- rxfinHallReq
			}
		case newrxLocalState := <-RecieveLocalStateRx:
			if rxLocalState != newrxLocalState {
				rxLocalState = newrxLocalState
				fmt.Printf("NET:LocalState:: %+v\n", rxLocalState)
				RxElevInfoChan <- rxLocalState
			}
		case newrxnewOrder := <-RecieveOrderRx:
			//if !reflect.DeepEqual(rxnewOrder, newrxnewOrder) {
			rxnewOrder = newrxnewOrder
			fmt.Printf("NET:NewOrder:: %+v\n", rxnewOrder)
			RxNewOrdersChan <- rxnewOrder
			//}
		}
	}
}

func printPeerUpdate(p peers.PeerUpdate) {
	fmt.Printf("Peer update:\n")
	fmt.Printf(" Peers:  %q\n", p.Peers)
	fmt.Printf(" New: %q\n", p.New)
	fmt.Printf(" Lost: %q\n", p.Lost)
}
