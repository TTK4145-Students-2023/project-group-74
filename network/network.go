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
	port6         = 21000
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
	RxP2PElevInfoChan chan<- localTypes.P2P_ELEV_INFO,
	TxHRAInputChan <-chan localTypes.HRAInput,
	RxHRAInputChan chan<- localTypes.HRAInput,
	LostElevChan chan<- []string) {

	var (
		p2pElevInfo_Tx = localTypes.P2P_ELEV_INFO{}
		HallReq_Tx     = localTypes.BUTTON_INFO{Floor: localTypes.NUM_FLOORS, Button: localTypes.Button_Cab}
		finHallReq_Tx  = localTypes.BUTTON_INFO{Floor: localTypes.NUM_FLOORS, Button: localTypes.Button_Cab}
		localState_Tx  = localTypes.LOCAL_ELEVATOR_INFO{}
		newOrder_Tx    = map[string]localTypes.HMATRIX{}
		newHRAInput_Tx = localTypes.HRAInput{}

		p2pElevInfo_Rx = localTypes.P2P_ELEV_INFO{}
		HallReq_Rx     = localTypes.BUTTON_INFO{Floor: localTypes.NUM_FLOORS, Button: localTypes.Button_Cab}
		finHallReq_Rx  = localTypes.BUTTON_INFO{Floor: localTypes.NUM_FLOORS, Button: localTypes.Button_Cab}
		LocalState_Rx  = localTypes.LOCAL_ELEVATOR_INFO{}
		newOrder_Rx    = map[string]localTypes.HMATRIX{}
		newHRAInput_Rx = localTypes.HRAInput{}

		BCLocalState     = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		BCNewHallRequest = make(chan localTypes.BUTTON_INFO)
		BCFinHallOrder   = make(chan localTypes.BUTTON_INFO)
		BCNewOrder       = make(chan map[string]localTypes.HMATRIX)
		BCP2PElevInfo    = make(chan localTypes.P2P_ELEV_INFO)
		BCNewHRAInput    = make(chan localTypes.HRAInput)

		RecieveLocalState     = make(chan localTypes.LOCAL_ELEVATOR_INFO)
		RecieveNewHallRequest = make(chan localTypes.BUTTON_INFO)
		RecieveFinHallOrder   = make(chan localTypes.BUTTON_INFO)
		RecieveOrder          = make(chan map[string]localTypes.HMATRIX)
		RecieveP2PElevInfo    = make(chan localTypes.P2P_ELEV_INFO)
		RecieveNewHRAInput    = make(chan localTypes.HRAInput)
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

	// GoRoutines to broadcast(Tx) over NTW
	go bcast.Transmitter(Port1, BCLocalState)
	go bcast.Transmitter(Port2, BCNewHallRequest)
	go bcast.Transmitter(port3, BCFinHallOrder)
	go bcast.Transmitter(port4, BCNewOrder)
	go bcast.Transmitter(port5, BCP2PElevInfo)
	go bcast.Transmitter(port6, BCNewHRAInput)

	// GoRoutines to recieve(Rx) from NTW
	go bcast.Receiver(Port1, RecieveLocalState)
	go bcast.Receiver(Port2, RecieveNewHallRequest)
	go bcast.Receiver(port3, RecieveFinHallOrder)
	go bcast.Receiver(port4, RecieveOrder)
	go bcast.Receiver(port5, RecieveP2PElevInfo)
	go bcast.Receiver(port6, RecieveNewHRAInput)

	recieveTimer := time.NewTimer(1)
	recieveTimer.Stop()

	for {
		select {
		case p := <-peerUpdateCh:
			printPeerUpdate(p)
			PeerList.Peers = p.Peers
			if len(p.Lost) != 0 {
				LostElevChan <- p.Lost
			}
		case localState_Tx = <-TxElevInfoChan:
			BCLocalState <- localState_Tx
		case HallReq_Tx = <-TxNewHallRequestChan:
			BCNewHallRequest <- HallReq_Tx
		case finHallReq_Tx = <-TxFinishedHallOrderChan:
			BCFinHallOrder <- finHallReq_Tx
		case newOrder_Tx = <-TxNewOrdersChan:
			BCNewOrder <- newOrder_Tx
		case p2pElevInfo_Tx = <-TxP2PElevInfoChan:
			BCP2PElevInfo <- p2pElevInfo_Tx
		case newHRAInput_Tx = <-TxHRAInputChan:
			BCNewHRAInput <- newHRAInput_Tx

		case p2pElevInfo_Rx = <-RecieveP2PElevInfo:
			RxP2PElevInfoChan <- p2pElevInfo_Rx
		case newHallReq_Rx := <-RecieveNewHallRequest:
			if newHallReq_Rx.Floor != localTypes.NUM_FLOORS {
				HallReq_Rx = newHallReq_Rx
				RxNewHallRequestChan <- HallReq_Rx
			}
		case newFinHallReq := <-RecieveFinHallOrder:
			if newFinHallReq.Floor != localTypes.NUM_FLOORS {
				finHallReq_Rx = newFinHallReq
				RxFinishedHallOrderChan <- finHallReq_Rx
			}
		case newLocalState_Rx := <-RecieveLocalState:
			if LocalState_Rx != newLocalState_Rx {
				LocalState_Rx = newLocalState_Rx
				RxElevInfoChan <- LocalState_Rx
			}
		case newOrder_Rx = <-RecieveOrder:
			RxNewOrdersChan <- newOrder_Rx
		case newHRAInput_Rx = <-RecieveNewHRAInput:
			RxHRAInputChan <- newHRAInput_Rx
		}
	}
}

// -----  PUBLIC FUNCTIONS (NETWORK) ------ //

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

func printPeerUpdate(p peers.PeerUpdate) {
	fmt.Printf("Peer update:\n")
	fmt.Printf(" Peers:  %q\n", p.Peers)
	fmt.Printf(" New: %q\n", p.New)
	fmt.Printf(" Lost: %q\n", p.Lost)
}
