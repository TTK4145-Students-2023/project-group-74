package network

import (
	"fmt"
	"net"
	"project-group-74/localTypes"
	"project-group-74/network/subs/bcast"
	"project-group-74/network/subs/localip"
	"project-group-74/network/subs/peers"
	"time"
)

// ----- CONSTANTS (NETWORK) ------ //
const (
	PeerPort      = 15699
	Port1         = 16599
	Port2         = 17000
	port3         = 18000
	port4         = 19000
	port5         = 20000
	BroadcastRate = 100 * time.Millisecond
)

// ----- VARIABLES (NETWORK) ------ //
var MyIP string

var PeerList peers.PeerUpdate

// ----- MAIN FUNCTION (NETWORK) ------ //
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
		p2pElevInfo = localTypes.P2P_ELEV_INFO{}
		newHallReq  = localTypes.BUTTON_INFO{Floor: 4, Button: localTypes.Button_Cab}
		finHallReq  = localTypes.BUTTON_INFO{Floor: 4, Button: localTypes.Button_Cab}
		localState  = localTypes.LOCAL_ELEVATOR_INFO{}
		newOrder    = map[string]localTypes.HMATRIX{}

		rxP2pElevinfo = localTypes.P2P_ELEV_INFO{}
		rxnewHallReq  = localTypes.BUTTON_INFO{Floor: 4, Button: localTypes.Button_Cab}
		rxfinHallReq  = localTypes.BUTTON_INFO{Floor: 4, Button: localTypes.Button_Cab}
		rxLocalState  = localTypes.LOCAL_ELEVATOR_INFO{}
		rxnewOrder    = map[string]localTypes.HMATRIX{}

		BCLocalStateTx   = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		BCNewHallReqTx   = make(chan localTypes.BUTTON_INFO)
		BCFinHallOrderTx = make(chan localTypes.BUTTON_INFO)
		BCNewOrderTx     = make(chan map[string]localTypes.HMATRIX)
		BCP2PElevInfoTx  = make(chan localTypes.P2P_ELEV_INFO)

		RecieveLocalStateRx   = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		RecieveNewHallReqRx   = make(chan localTypes.BUTTON_INFO)
		RecieveFinHallOrderRx = make(chan localTypes.BUTTON_INFO)
		RecieveOrderRx        = make(chan map[string]localTypes.HMATRIX)
		RecieveP2PElevInfo    = make(chan localTypes.P2P_ELEV_INFO)
	)

	if MyIP == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "No IP available"
		}
		MyIP = localIP
	}
	fmt.Printf(" NETWORK RUNNING\n")

	peerUpdateCh := make(chan peers.PeerUpdate) //channel for receiving IDs of alive peers
	peerTxEnable := make(chan bool)

	go peers.Transmitter(PeerPort, MyIP, peerTxEnable)
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

	recieveTimer := time.NewTimer(1)
	recieveTimer.Stop()

	for {
		select {
		case p := <-peerUpdateCh:
			printPeerUpdate(p)
			PeerList.Peers = p.Peers
		case localState = <-TxElevInfoChan:
			//fmt.Printf("NET: Transmit local state::   %+v\n", localState)
			BCLocalStateTx <- localState
		case newHallReq = <-TxNewHallRequestChan:
			//fmt.Printf("NET: Transmit new hall req::   %+v\n", rxnewHallReq)
			BCNewHallReqTx <- newHallReq
		case finHallReq = <-TxFinishedHallOrderChan:
			//fmt.Printf("NET: Transmit finished hall order::   %+v\n", finHallReq)
			BCFinHallOrderTx <- finHallReq
		case newOrder = <-TxNewOrdersChan:
			//fmt.Printf("NET: Transmit new order::   %+v\n", newOrder)
			BCNewOrderTx <- newOrder
		case p2pElevInfo = <-TxP2PElevInfoChan:
			//fmt.Printf("NET: Transmit P2Pelevinfo::   %+v\n", p2pElevInfo)
			BCP2PElevInfoTx <- p2pElevInfo

		case newrxP2pElevinfo := <-RecieveP2PElevInfo:
			/*sort.Slice(rxP2pElevinfo, func(i, j int) bool {
				return rxP2pElevinfo[i].ElevID < rxP2pElevinfo[j].ElevID
			})
			sort.Slice(newrxP2pElevinfo, func(i, j int) bool {
				return newrxP2pElevinfo[i].ElevID < newrxP2pElevinfo[j].ElevID
			})
			if !reflect.DeepEqual(rxP2pElevinfo, newrxP2pElevinfo) {*/
			rxP2pElevinfo = newrxP2pElevinfo
			//fmt.Printf("NET:P2Pelevinfo:: %+v\n", rxP2pElevinfo)
			RxP2PElevInfoChan <- rxP2pElevinfo
			//}
		case newrxnewHallReq := <-RecieveNewHallReqRx:
			if rxnewHallReq != newrxnewHallReq {
				rxnewHallReq = newrxnewHallReq
				RxNewHallRequestChan <- rxnewHallReq
			}
		case newrxfinHallReq := <-RecieveFinHallOrderRx:
			if rxfinHallReq != newrxfinHallReq {
				rxfinHallReq = newrxfinHallReq
				RxFinishedHallOrderChan <- rxfinHallReq
			}
		case newrxLocalState := <-RecieveLocalStateRx:
			if rxLocalState != newrxLocalState {
				rxLocalState = newrxLocalState
				RxElevInfoChan <- rxLocalState
			}
		case newrxnewOrder := <-RecieveOrderRx:
			rxnewOrder = newrxnewOrder
			RxNewOrdersChan <- rxnewOrder
		}
	}
}

// -----  PUBLIC FUNCTIONS (NETWORK) ------ //
func printPeerUpdate(p peers.PeerUpdate) {
	fmt.Printf("Peer update:\n")
	fmt.Printf(" Peers:  %q\n", p.Peers)
	fmt.Printf(" New: %q\n", p.New)
	fmt.Printf(" Lost: %q\n", p.Lost)
}

func IsMaster(MyIP string, Peers []string) bool {
	if len(Peers) == 0 {
		return true
	}
	lowestIP := Peers[0]
	for _, ip := range Peers {
		lastOctet := splitIPAddr(ip)
		addrLowest := net.ParseIP(lowestIP).To4()
		if lastOctet < addrLowest[3] {
			lowestIP = ip
		}
	}
	myIP := net.ParseIP(MyIP).To4()
	lowestIP = string(net.ParseIP(lowestIP).To4())
	return myIP[3] <= lowestIP[3]
}

func SendlocalElevInfo(MyElev localTypes.LOCAL_ELEVATOR_INFO, RXchan chan<- localTypes.LOCAL_ELEVATOR_INFO, TXchan chan<- localTypes.LOCAL_ELEVATOR_INFO) {
	if len(PeerList.Peers) == 0 {
		RXchan <- MyElev
	} else {
		TXchan <- MyElev
	}
}

func SendNewOrders(orders map[string]localTypes.HMATRIX, RXButtonchan chan<- map[string]localTypes.HMATRIX, TXButtonchan chan<- map[string]localTypes.HMATRIX) {
	if len(PeerList.Peers) == 0 {
		RXButtonchan <- orders
	} else {
		TXButtonchan <- orders
	}
}

func SendButtonInfo(MyElev localTypes.LOCAL_ELEVATOR_INFO, btntype localTypes.BUTTON_TYPE, RXButtonchan chan<- localTypes.BUTTON_INFO, TXButtonchan chan<- localTypes.BUTTON_INFO) {
	if len(PeerList.Peers) == 0 {
		RXButtonchan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: btntype}
	} else {
		TXButtonchan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: btntype}
	}
}

func SendButtonPress(MyElev localTypes.LOCAL_ELEVATOR_INFO, btnpress localTypes.BUTTON_INFO, RXButtonchan chan<- localTypes.BUTTON_INFO, TXButtonchan chan<- localTypes.BUTTON_INFO) {
	if len(PeerList.Peers) == 0 {
		RXButtonchan <- btnpress
	} else {
		TXButtonchan <- btnpress
	}
}

// -----  PRIVATE FUNCTIONS (NETWORK) ------ //
func splitIPAddr(ip string) byte {
	addr := net.ParseIP(ip).To4()
	return addr[3]
}
