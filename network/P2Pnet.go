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
	Port1         = 16599
	Port2         = 17000
	port3         = 18000
	port4         = 19000
	port5         = 20000
	port6         = 21000
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
	RxP2PElevInfoChan chan<- localTypes.P2P_ELEV_INFO,
	TxHRAInputChan <-chan localTypes.HRAInput,
	RxHRAInputChan chan<- localTypes.HRAInput,
	LostElevChan chan<- []string) {

	var (
		// Tx var
		p2pElevInfo = localTypes.P2P_ELEV_INFO{}
		newHallReq  = localTypes.BUTTON_INFO{Floor: 4, Button: localTypes.Button_Cab}
		finHallReq  = localTypes.BUTTON_INFO{Floor: 4, Button: localTypes.Button_Cab}
		localState  = localTypes.LOCAL_ELEVATOR_INFO{}
		newOrder    = map[string]localTypes.HMATRIX{}
		newHRAInput = localTypes.HRAInput{}
		// Rx var
		rxP2pElevinfo = localTypes.P2P_ELEV_INFO{}
		rxnewHallReq  = localTypes.BUTTON_INFO{Floor: 4, Button: localTypes.Button_Cab}
		rxfinHallReq  = localTypes.BUTTON_INFO{Floor: 4, Button: localTypes.Button_Cab}
		rxLocalState  = localTypes.LOCAL_ELEVATOR_INFO{}
		rxnewOrder    = map[string]localTypes.HMATRIX{}
		rxnewHRAInput = localTypes.HRAInput{}

		// Tx chan
		BCLocalStateTx   = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		BCNewHallReqTx   = make(chan localTypes.BUTTON_INFO)
		BCFinHallOrderTx = make(chan localTypes.BUTTON_INFO)
		BCNewOrderTx     = make(chan map[string]localTypes.HMATRIX)
		BCP2PElevInfoTx  = make(chan localTypes.P2P_ELEV_INFO)
		BCNewHRAInput    = make(chan localTypes.HRAInput)
		// Rx chan
		RecieveLocalStateRx   = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		RecieveNewHallReqRx   = make(chan localTypes.BUTTON_INFO)
		RecieveFinHallOrderRx = make(chan localTypes.BUTTON_INFO)
		RecieveOrderRx        = make(chan map[string]localTypes.HMATRIX)
		RecieveP2PElevInfo    = make(chan localTypes.P2P_ELEV_INFO)
		RecieveNewHRAInput    = make(chan localTypes.HRAInput)
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
	go bcast.Receiver(port6, RecieveNewHRAInput)

	// GoRoutines to broadcast over NTW
	go bcast.Transmitter(Port1, BCLocalStateTx)
	go bcast.Transmitter(Port2, BCNewHallReqTx)
	go bcast.Transmitter(port3, BCFinHallOrderTx)
	go bcast.Transmitter(port4, BCNewOrderTx)
	go bcast.Transmitter(port5, BCP2PElevInfoTx)
	go bcast.Transmitter(port6, BCNewHRAInput)

	// Broadcast Timer
	//broadcastTimer := time.NewTimer(BroadcastRate)
	recieveTimer := time.NewTimer(1)
	recieveTimer.Stop()

	for {
		select {
		// case der vi sjekker om vi er i init state, og tømmer? variablene

		// Print Peer Updates
		case p := <-peerUpdateCh:
			printPeerUpdate(p)
			localTypes.PeerList.Peers = p.Peers
			if len(p.Lost) != 0 {
				LostElevChan <- p.Lost
			}
			// Broadcasting on network
		case localState = <-TxElevInfoChan:
			BCLocalStateTx <- localState
		case newHallReq = <-TxNewHallRequestChan:

			BCNewHallReqTx <- newHallReq

		case finHallReq = <-TxFinishedHallOrderChan:
			BCFinHallOrderTx <- finHallReq
		case newOrder = <-TxNewOrdersChan:
			BCNewOrderTx <- newOrder
		case p2pElevInfo = <-TxP2PElevInfoChan:
			BCP2PElevInfoTx <- p2pElevInfo
		case newHRAInput = <-TxHRAInputChan:
			BCNewHRAInput <- newHRAInput
		/*case <-broadcastTimer.C:
		fmt.Printf("NET: Broadcasting NOW\n")
		BCLocalStateTx <- localState
		BCNewHallReqTx <- newHallReq
		BCFinHallOrderTx <- finHallReq
		BCNewOrderTx <- newOrder
		BCP2PElevInfoTx <- p2pElevInfo
		broadcastTimer.Reset(BroadcastRate)*/

		case rxP2pElevinfo = <-RecieveP2PElevInfo: // Legge på sender ID?

			RxP2PElevInfoChan <- rxP2pElevinfo

		//	}
		case newrxnewHallReq := <-RecieveNewHallReqRx:

			if newrxnewHallReq.Floor != 4 {
				rxnewHallReq = newrxnewHallReq
				RxNewHallRequestChan <- rxnewHallReq
			}

		case newrxfinHallReq := <-RecieveFinHallOrderRx:
			if newrxfinHallReq.Floor != 4 {
				rxfinHallReq = newrxfinHallReq
				RxFinishedHallOrderChan <- rxfinHallReq
			}
		case newrxLocalState := <-RecieveLocalStateRx:
			if rxLocalState != newrxLocalState {
				rxLocalState = newrxLocalState
				RxElevInfoChan <- rxLocalState
			}
		case newrxnewOrder := <-RecieveOrderRx:
			//if !reflect.DeepEqual(rxnewOrder, newrxnewOrder) {
			rxnewOrder = newrxnewOrder
			RxNewOrdersChan <- rxnewOrder
			//}
		case rxnewHRAInput = <-RecieveNewHRAInput:
			RxHRAInputChan <- rxnewHRAInput

		}
	}
}

func printPeerUpdate(p peers.PeerUpdate) {
	fmt.Printf("Peer update:\n")
	fmt.Printf(" Peers:  %q\n", p.Peers)
	fmt.Printf(" New: %q\n", p.New)
	fmt.Printf(" Lost: %q\n", p.Lost)
}
