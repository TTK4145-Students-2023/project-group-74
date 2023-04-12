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

	var (
		// Tx var
		p2pElevInfo = localTypes.P2P_ELEV_INFO{}
		newHallReq  = localTypes.BUTTON_INFO{}
		finHallReq  = localTypes.BUTTON_INFO{}
		localState  = localTypes.LOCAL_ELEVATOR_INFO{}
		newOrder    = map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool{}
		// Rx var
		rxP2pElevinfo = localTypes.P2P_ELEV_INFO{}
		rxnewHallReq  = localTypes.BUTTON_INFO{}
		rxfinHallReq  = localTypes.BUTTON_INFO{}
		rxLocalState  = localTypes.LOCAL_ELEVATOR_INFO{}
		rxnewOrder    = map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool{}
		// Tx chan
		BCLocalStateTx   = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		BCNewHallReqTx   = make(chan localTypes.BUTTON_INFO)
		BCFinHallOrderTx = make(chan localTypes.BUTTON_INFO)
		BCNewOrderTx     = make(chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool)
		BCP2PElevInfoTx  = make(chan localTypes.P2P_ELEV_INFO)
		// Rx chan
		RecieveLocalStateRx   = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		RecieveNewHallReqRx   = make(chan localTypes.BUTTON_INFO)
		RecieveFinHallOrderRx = make(chan localTypes.BUTTON_INFO)
		RecieveOrderRx        = make(chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool)
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
	go bcast.Transmitter(StatePort, BCP2PElevInfoTx)

	// Broadcast Timer
	broadcastTimer := time.NewTimer(BroadcastRate)
	recieveTimer := time.NewTimer(1)
	recieveTimer.Stop()

	fmt.Printf("RecieveTimer: %v\n", recieveTimer)

	for {
		select {

		// Print Peer Updates
		case p := <-peerUpdateCh:
			printPeerUpdate(p)
			localTypes.PeerList.Peers = p.Peers
			// Broadcasting on network
		case localState = <-TxElevInfoChan:
		case newHallReq = <-TxNewHallRequestChan:
		case finHallReq = <-TxFinishedHallOrderChan:
		case newOrder = <-TxNewOrdersChan:
		case p2pElevInfo = <-TxP2PElevInfoChan:
		case <-broadcastTimer.C:
			fmt.Printf("NET.BC: Broadcasting\n")
			BCLocalStateTx <- localState
			BCNewHallReqTx <- newHallReq
			BCFinHallOrderTx <- finHallReq
			BCNewOrderTx <- newOrder
			BCP2PElevInfoTx <- p2pElevInfo
			broadcastTimer.Reset(BroadcastRate)

			// Reading from network
		case newrxP2pElevinfo := <-RecieveP2PElevInfo:
			fmt.Printf("NET.RX: P2PElevInfo\n")
			sort.Slice(rxP2pElevinfo, func(i, j int) bool {
				return rxP2pElevinfo[i].ElevID < rxP2pElevinfo[j].ElevID
			})
			sort.Slice(newrxP2pElevinfo, func(i, j int) bool {
				return newrxP2pElevinfo[i].ElevID < newrxP2pElevinfo[j].ElevID
			})
			if !reflect.DeepEqual(rxP2pElevinfo, newrxP2pElevinfo) {
				rxP2pElevinfo = newrxP2pElevinfo
				RxP2PElevInfoChan <- rxP2pElevinfo
			}
		case newrxnewHallReq := <-RecieveNewHallReqRx:
			fmt.Printf("NET.RX: newHallReq\n")
			if rxnewHallReq != newrxnewHallReq {
				rxnewHallReq = newrxnewHallReq
				RxNewHallRequestChan <- rxnewHallReq
			}
		case newrxfinHallReq := <-RecieveFinHallOrderRx:
			fmt.Printf("NET.RX: finHallreq\n")
			if rxfinHallReq != newrxfinHallReq {
				rxfinHallReq = newrxfinHallReq
				RxFinishedHallOrderChan <- rxfinHallReq
			}
		case newrxLocalState := <-RecieveLocalStateRx:
			fmt.Printf("NET.RX: localState\n")
			if rxLocalState != newrxLocalState {
				rxLocalState = newrxLocalState
				RxElevInfoChan <- rxLocalState
			}
		case newrxnewOrder := <-RecieveOrderRx:
			fmt.Printf("NET.RX: newOrder\n")
			if !reflect.DeepEqual(rxnewOrder, newrxnewOrder) {
				rxnewOrder = newrxnewOrder
				RxNewOrdersChan <- rxnewOrder
			}
		}
	}
}

func printPeerUpdate(p peers.PeerUpdate) {
	fmt.Printf("Peer update:\n")
	fmt.Printf(" Peers:  %q\n", p.Peers)
	fmt.Printf(" New: %q\n", p.New)
	fmt.Printf(" Lost: %q\n", p.Lost)
}
