package elevio

import (
	"fmt"
	"net"
	"project-group-74/localTypes"
	"sync"
	"time"
)

// ----- CONSTANTS (ELEVATOR IO)------ //
const _pollRate = 20 * time.Millisecond

// ----- VARIABLES (ELEVATOR IO)------ //
var _initialized bool = false
var _mtx sync.Mutex
var _conn net.Conn
var _numFloors = localTypes.NUM_FLOORS

// ----- PUBLIC FUNCTIONS (ELEVATOR IO)------ //
func Init(addr string, numFloors int) {

	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func PollButtons(receiver chan<- localTypes.BUTTON_INFO) {
	fmt.Printf(" POLLBUTTONS RUNNING \n")

	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)

		for f := 0; f < _numFloors; f++ {
			for b := localTypes.BUTTON_TYPE(0); b < 3; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v {
					receiver <- localTypes.BUTTON_INFO{Floor: f, Button: localTypes.BUTTON_TYPE(b)}
				}
				prev[f][b] = v
			}

		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	fmt.Printf(" POLLFLOOR RUNNING ")

	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
			fmt.Printf("New Floor: %v \n", v)
		}
		prev = v

	}

}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

// ----- PUBLIC GET/SET FUNCTIONS (ELEVATOR IO)------ //
func SetMotorDirection(dir localTypes.MOTOR_DIR) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button localTypes.BUTTON_TYPE, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func GetButton(button localTypes.BUTTON_TYPE, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

// ----- PRIVATE FUNCTIONS (ELEVATOR IO)------ //
func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
