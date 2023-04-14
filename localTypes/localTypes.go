package localTypes

import (
	//project config

	"net"
	"project-group-74/network/subs/peers"
	"strconv"
	"time"
)

// ----- CONSTANTS ------ //
// Create an init file with the following constants
// Time
// RX_BUFFER

const (
	NUM_BUTTONS = 3
	NUM_FLOORS  = 4
	NUM_ORDERS  = NUM_FLOORS * NUM_BUTTONS

	OPEN_DOOR_TIME_sek       = 3
	TRAVEL_TIME_sek          = 3
	MAX_TIME_TO_FINISH_ORDER = 3 * (NUM_FLOORS - 1) * (TRAVEL_TIME_sek * OPEN_DOOR_TIME_sek)
	P2P_UPDATE_INTERVAL      = 2000 //in ms
)

// ----- TYPE DEFINITIONS ------ //

type BUTTON_TYPE int

const (
	Button_Cab       BUTTON_TYPE = 2
	Button_hall_up               = 0
	Button_hall_down             = 1
)

type BUTTON_INFO struct {
	Floor  int
	Button BUTTON_TYPE
}

type HMATRIX [NUM_FLOORS][NUM_BUTTONS - 1]bool
type ORDER map[string]HMATRIX
type P2P_ELEV_INFO []LOCAL_ELEVATOR_INFO

type FOREIGN_ORDER_TYPE struct {
	Foreign_order BUTTON_INFO
	Active        bool
	Local         bool
}

type ELEVATOR_STATE int

const (
	Idle      ELEVATOR_STATE = 0
	Moving                   = 1
	Door_open                = 2
)

type LOCAL_ELEVATOR_INFO struct {
	Floor     int
	Direction MOTOR_DIR
	State     ELEVATOR_STATE
	CabCalls  [NUM_FLOORS]bool 
	ElevID    string
}

type MOTOR_DIR int

const (
	DIR_down MOTOR_DIR = -1
	DIR_stop           = 0
	DIR_up             = 1
)

const ORDER_WATCHDOG_POLL_RATE = 50 * time.Millisecond

type HRAElevState struct {
	State       string           `json:"behaviour"`
	Floor       int              `json:"floor"`
	Direction   string           `json:"direction"`
	CabRequests [NUM_FLOORS]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [NUM_FLOORS][2]bool     `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

type orderAssignerBehavior int

const (
	OABehaviorMaster orderAssignerBehavior = iota
	OABehaviorSlave
)

// ----- FUNCTIONS (VALIDATION) ------ //
func isValidFloor(floor int) bool {
	return floor >= 0 && floor <= NUM_FLOORS
}

func isValidID(ID string) bool {
	id, err := strconv.Atoi(ID)
	if err != nil || id < 0 {
		return false
	}
	return true
}

func (state ELEVATOR_STATE) isValid() bool {
	return state == Idle ||
		state == Moving ||
		state == Door_open
}

func (button BUTTON_TYPE) isValid() bool {
	return button == Button_Cab ||
		button == Button_hall_up ||
		button == Button_hall_down
}

func (btnInfo BUTTON_INFO) isValid() bool {
	return btnInfo.Button.isValid() && isValidFloor(btnInfo.Floor)
}

func (order FOREIGN_ORDER_TYPE) isValid() bool {
	return BUTTON_INFO(order.Foreign_order).isValid()
}

func (dir MOTOR_DIR) isValid() bool {
	return dir == DIR_down ||
		dir == DIR_up ||
		dir == DIR_stop
}

func (loc_elev LOCAL_ELEVATOR_INFO) isValid() bool {
	return isValidFloor(loc_elev.Floor) &&
		loc_elev.Direction.isValid() &&
		loc_elev.State.isValid()
}

//************ const for P2P ************

const (
	PeerPort  = 15699
	StatePort = 16599
)

var MyIP string

var PeerList peers.PeerUpdate

// ----- FUNCTIONS (NETWORK) ------ //
func splitIPAddr(ip string) byte {
	addr := net.ParseIP(ip).To4()
	return addr[3]
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
	/*fmt.Printf("My IP: %v\n", myIP)
	v := myIP[3] <= lowestIP[3]
	fmt.Printf("Am I master: %v\n", v)*/
	return myIP[3] <= lowestIP[3]
}

func SendlocalElevInfo(MyElev LOCAL_ELEVATOR_INFO, RXchan chan<- LOCAL_ELEVATOR_INFO, TXchan chan<- LOCAL_ELEVATOR_INFO){
	if len(PeerList.Peers) == 0 {
		RXchan <- MyElev
	} else {
		TXchan <- MyElev
	}
}

func SendButtonInfo(MyElev LOCAL_ELEVATOR_INFO, btntype BUTTON_TYPE, RXButtonchan chan<- BUTTON_INFO, TXButtonchan chan<- BUTTON_INFO){
	if len(PeerList.Peers) == 0 {
		RXButtonchan <- BUTTON_INFO{Floor: MyElev.Floor, Button: btntype}
	} else {
		TXButtonchan <- BUTTON_INFO{Floor: MyElev.Floor, Button: btntype}
	}
}