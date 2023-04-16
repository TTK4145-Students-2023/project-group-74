package localTypes

import (
	"strconv"
	"strings"
	"time"
)

// ----- CONSTANTS ------ //
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
func IsValidFloor(floor int) bool {
	return floor >= 0 && floor <= NUM_FLOORS
}

func IsValidID(ip string) bool {
	octets := strings.Split(ip, ".")
	if len(octets) != 4 {
		return false
	}
	for _, octet := range octets {
		num, err := strconv.Atoi((octet))
		if err != nil || num < 0 || num > 255 {
			return false
		}
	}
	return true
}

func (state ELEVATOR_STATE) IsValid() bool {
	return state == Idle ||
		state == Moving ||
		state == Door_open
}

func (button BUTTON_TYPE) IsValid() bool {
	return button == Button_Cab ||
		button == Button_hall_up ||
		button == Button_hall_down
}

func (btnInfo BUTTON_INFO) IsValid() bool {
	return btnInfo.Button.IsValid() && IsValidFloor(btnInfo.Floor)
}

func (order FOREIGN_ORDER_TYPE) IsValid() bool {
	return BUTTON_INFO(order.Foreign_order).IsValid()
}

func (dir MOTOR_DIR) IsValid() bool {
	return dir == DIR_down ||
		dir == DIR_up ||
		dir == DIR_stop
}

func (loc_elev LOCAL_ELEVATOR_INFO) IsValid() bool {
	return IsValidFloor(loc_elev.Floor) &&
		loc_elev.Direction.IsValid() &&
		loc_elev.State.IsValid()
}
