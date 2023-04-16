package elevio

import (
	"project-group-74/localTypes"
	"time"
)

// ----- PUBLIC FUNCTIONS (ELEVATOR FUNCTIONS)------ //
func ArrivedAtOrder(
	MyElev localTypes.LOCAL_ELEVATOR_INFO) localTypes.LOCAL_ELEVATOR_INFO {

	MyElev.Direction = localTypes.DIR_stop
	MyElev.State = localTypes.Door_open
	SetMotorDirection(localTypes.DIR_stop)
	SetDoorOpenLamp(true)

	doorTimer := time.NewTimer(3 * time.Second)
	<-doorTimer.C
	doorTimer.Stop()
	SetDoorOpenLamp(false)
	MyElev.State = localTypes.Idle

	return MyElev
}

func SendWithDelay(foreignElevs localTypes.P2P_ELEV_INFO, TxChannel chan<- localTypes.P2P_ELEV_INFO) {
	timer := time.NewTimer(localTypes.P2P_UPDATE_INTERVAL * time.Millisecond)
	<-timer.C
	TxChannel <- foreignElevs
}

func IsHOrderActive(newOrder localTypes.BUTTON_INFO, CurrentHMatrix localTypes.HMATRIX) bool {
	return CurrentHMatrix[newOrder.Floor][newOrder.Button]
}

func IsOrderAtFloor(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	if MyElev.CabCalls[MyElev.Floor] || MyOrders[MyElev.Floor][localTypes.Button_hall_down] || MyOrders[MyElev.Floor][localTypes.Button_hall_up] {
		return true
	} else {
		return false
	}
}

func AddLocalToForeignInfo(MyElev localTypes.LOCAL_ELEVATOR_INFO, ForeignElevs localTypes.P2P_ELEV_INFO) localTypes.P2P_ELEV_INFO {
	for i := 0; i < len(ForeignElevs); i++ {
		if ForeignElevs[i].ElevID == MyElev.ElevID {
			ForeignElevs[i] = MyElev
		}
	}
	return ForeignElevs
}

func UpdateOrderLights(MyElev localTypes.LOCAL_ELEVATOR_INFO, CurrentHMatrix localTypes.HMATRIX) {
	for f := 0; f < localTypes.NUM_FLOORS; package DLOCC

	import (
		"project-group-74/localTypes"
		"time"
	)
	
	const ORDER_WATCHDOG_POLL_RATE = 50 * time.Millisecond
	
	
	var motorDirStrings = map[localTypes.MOTOR_DIR]string{
		localTypes.DIR_down: "down",
		localTypes.DIR_stop: "stop",
		localTypes.DIR_up:   "up",
	}
	
	var elevStateStrings = map[localTypes.ELEVATOR_STATE]string{
		localTypes.Idle:      "idle",
		localTypes.Moving:    "moving",
		localTypes.Door_open: "doorOpen",
	}
	f++ {
		SetButtonLamp(localTypes.Button_Cab, f, MyElev.CabCalls[f])
		for btn := 0; btn < localTypes.NUM_BUTTONS-1; btn++ {
			SetButtonLamp(localTypes.BUTTON_TYPE(btn), f, CurrentHMatrix[f][btn])
		}
	}
}

func LocalElevInitFloor(MyElev localTypes.LOCAL_ELEVATOR_INFO) localTypes.LOCAL_ELEVATOR_INFO {
	for GetFloor() == -1 {
		SetMotorDirection(localTypes.DIR_down)
	}
	SetMotorDirection(localTypes.DIR_stop)
	MyElev.Floor = GetFloor()

	return MyElev
}

func RemoveOneOrderBtn(finishedOrder localTypes.BUTTON_INFO, MyElev localTypes.LOCAL_ELEVATOR_INFO) [4]bool {
	MyElev.CabCalls[finishedOrder.Floor] = false
	return MyElev.CabCalls
}

func AddOneNewOrderBtn(newOrder localTypes.BUTTON_INFO, MyElev localTypes.LOCAL_ELEVATOR_INFO) [4]bool { //neccecary?
	MyElev.CabCalls[newOrder.Floor] = true
	return MyElev.CabCalls
}

func FindDirection(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) (localTypes.MOTOR_DIR, localTypes.ELEVATOR_STATE) {
	switch true {
	case Requests_here(MyElev, MyOrders):
		return localTypes.DIR_stop, localTypes.Door_open
	case MyElev.Direction == localTypes.DIR_up && Requests_above(MyElev, MyOrders):
		return localTypes.DIR_up, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_up && Requests_below(MyElev, MyOrders):
		return localTypes.DIR_down, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_down && Requests_below(MyElev, MyOrders):
		return localTypes.DIR_down, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_down && Requests_above(MyElev, MyOrders):
		return localTypes.DIR_up, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_stop && Requests_above(MyElev, MyOrders):
		return localTypes.DIR_up, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_stop && Requests_below(MyElev, MyOrders):
		return localTypes.DIR_down, localTypes.Moving
	default:
		return localTypes.DIR_stop, localTypes.Idle
	}
}

func FindDirectionNotHere(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) (localTypes.MOTOR_DIR, localTypes.ELEVATOR_STATE) {
	switch true {
	case MyElev.Direction == localTypes.DIR_up && Requests_above(MyElev, MyOrders):
		return localTypes.DIR_up, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_up && Requests_below(MyElev, MyOrders):
		return localTypes.DIR_down, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_up && Requests_here(MyElev, MyOrders):
		return localTypes.DIR_stop, localTypes.Door_open
	case MyElev.Direction == localTypes.DIR_down && Requests_below(MyElev, MyOrders):
		return localTypes.DIR_down, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_down && Requests_above(MyElev, MyOrders):
		return localTypes.DIR_up, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_down && Requests_here(MyElev, MyOrders):
		return localTypes.DIR_stop, localTypes.Door_open
	case MyElev.Direction == localTypes.DIR_stop && Requests_above(MyElev, MyOrders):
		return localTypes.DIR_up, localTypes.Moving
	case MyElev.Direction == localTypes.DIR_stop && Requests_below(MyElev, MyOrders):
		return localTypes.DIR_down, localTypes.Moving

	default:
		return localTypes.DIR_stop, localTypes.Idle
	}
}

func Requests_here(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	totalOrders := combineOrders(MyElev.CabCalls, MyOrders)
	for btn := 0; btn < localTypes.NUM_BUTTONS; btn++ {
		if totalOrders[MyElev.Floor][btn] {
			return true
		}
	}
	return false
}

func Requests_above(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	totalOrders := combineOrders(MyElev.CabCalls, MyOrders)
	for f := MyElev.Floor + 1; f < localTypes.NUM_FLOORS; f++ {
		for btn := 0; btn < localTypes.NUM_BUTTONS; btn++ {
			if totalOrders[f][btn] {
				return true
			}
		}
	}
	return false
}

func Requests_below(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	totalOrders := combineOrders(MyElev.CabCalls, MyOrders)
	for f := 0; f < MyElev.Floor; f++ {
		for btn := 0; btn < localTypes.NUM_BUTTONS; btn++ {
			if totalOrders[f][btn] {
				return true
			}
		}
	}
	return false
}

func AddNewOrdersToLocal(newOrder localTypes.ORDER, MyOrders localTypes.HMATRIX, MyElev localTypes.LOCAL_ELEVATOR_INFO) localTypes.HMATRIX {
	for f := 0; f < localTypes.NUM_FLOORS; f++ {
		for btn := 0; btn < localTypes.NUM_BUTTONS-1; btn++ {
			(MyOrders)[f][btn] = newOrder[MyElev.ElevID][f][btn]
		}
	}
	return MyOrders
}

func AddNewOrdersToHMatrix(newOrder localTypes.ORDER) localTypes.HMATRIX {
	var CombinedHMatrix localTypes.HMATRIX
	for ID := range newOrder {
		for f := 0; f < localTypes.NUM_FLOORS; f++ {
			for btn := 0; btn < localTypes.NUM_BUTTONS-1; btn++ {
				if !CombinedHMatrix[f][btn] {
					(CombinedHMatrix)[f][btn] = newOrder[ID][f][btn]
				}
			}
		}
	}
	return CombinedHMatrix
}

// ----- PRIVATE FUNCTIONS (ELEVATOR FUNCTIONS)------ //
func combineOrders(MyCabs [localTypes.NUM_FLOORS]bool, MyOrders localTypes.HMATRIX) [localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS]bool {
	var result [localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS]bool
	for i := 0; i < localTypes.NUM_FLOORS; i++ {
		result[i][0] = MyCabs[i]
		for j := 1; j < localTypes.NUM_BUTTONS; j++ {
			result[i][j] = MyOrders[i][j-1]
		}
	}
	return result
}
