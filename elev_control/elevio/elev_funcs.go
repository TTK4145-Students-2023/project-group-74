package elevio

import (
	"project-group-74/localTypes"
	"time"
)

// used as goroutine

func ArrivedAtOrder(
	MyElev *localTypes.LOCAL_ELEVATOR_INFO) {

	MyElev.Direction = localTypes.DIR_stop
	MyElev.Floor = GetFloor()
	MyElev.State = localTypes.Door_open
	SetMotorDirection(localTypes.DIR_stop)
	SetDoorOpenLamp(true)

	doorTimer := time.NewTimer(3 * time.Second)
	<-doorTimer.C
	doorTimer.Stop()
	SetDoorOpenLamp(false)
	MyElev.State = localTypes.Idle
}

func SendWithDelay(foreignElevs localTypes.P2P_ELEV_INFO, TxChannel chan<- localTypes.P2P_ELEV_INFO) {
	timer := time.NewTimer(localTypes.P2P_UPDATE_INTERVAL * time.Millisecond)
	<-timer.C
	TxChannel <- foreignElevs
}

// Private funcs
func ChooseDirectionAndState(MyElev *localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) {
	newDir, newState := findDirection(MyElev, MyOrders)
	SetMotorDirection(newDir)
	MyElev.Direction = newDir
	MyElev.State = newState
}

func IsHOrderActive(newOrder localTypes.BUTTON_INFO, CurrentHMatrix localTypes.HMATRIX) bool { //neccecary?
	return CurrentHMatrix[newOrder.Floor][newOrder.Button]
}

func IsOrderAtFloor(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	btntype := dir2Btntype(MyElev.Direction)
	if MyElev.CabCalls[GetFloor()] || MyOrders[GetFloor()][btntype] {
		return true
	}
	return false
}

func AddNewOrders(newOrder localTypes.ORDER, MyOrders *localTypes.HMATRIX, CombinedHMatrix *localTypes.HMATRIX, MyElev localTypes.LOCAL_ELEVATOR_INFO) {
	addNewOrdersToLocal(newOrder, MyOrders, MyElev)
	addNewOrdersToHMatrix(newOrder, CombinedHMatrix)
}

func AddLocalToForeignInfo(MyElev localTypes.LOCAL_ELEVATOR_INFO, ForeignElevs *localTypes.P2P_ELEV_INFO) {
	for ForeignElev := 0; ForeignElev < len(*ForeignElevs); ForeignElev++ {
		if (*ForeignElevs)[ForeignElev].ElevID == MyElev.ElevID {
			(*ForeignElevs)[ForeignElev] = MyElev
		}
	}
}

func UpdateOrderLights(MyElev localTypes.LOCAL_ELEVATOR_INFO, CurrentHMatrix localTypes.HMATRIX) {
	for f := 0; f < localTypes.NUM_FLOORS; f++ {
		SetButtonLamp(localTypes.Button_Cab, f, MyElev.CabCalls[f])
		for btn := 0; btn < localTypes.NUM_BUTTONS-1; btn++ {
			SetButtonLamp(localTypes.BUTTON_TYPE(btn), f, CurrentHMatrix[f][btn])
		}
	}
}

func LocalElevInitFloor(MyElev *localTypes.LOCAL_ELEVATOR_INFO) {
	for GetFloor() == -1 {
		SetMotorDirection(localTypes.DIR_down)
	}
	SetMotorDirection(localTypes.DIR_stop)
	MyElev.Floor = GetFloor()
}

func GetFinOrder(floor int, pastDir localTypes.MOTOR_DIR) localTypes.BUTTON_INFO {
	btn := dir2Btntype(pastDir)
	btninfo := localTypes.BUTTON_INFO{
		Button: btn,
		Floor:  floor,
	}
	return btninfo
}

func RemoveOneOrderBtn(finishedOrder localTypes.BUTTON_INFO, MyElev *localTypes.LOCAL_ELEVATOR_INFO) {
	MyElev.CabCalls[finishedOrder.Floor] = false
}

func AddOneNewOrderBtn(newOrder localTypes.BUTTON_INFO, MyElev *localTypes.LOCAL_ELEVATOR_INFO) { //neccecary?
	MyElev.CabCalls[newOrder.Floor] = true
}

//Internal funcs

func findDirection(MyElev *localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) (localTypes.MOTOR_DIR, localTypes.ELEVATOR_STATE) {
	switch {
	case MyElev.Direction == localTypes.DIR_up:
		if requests_above(MyElev, MyOrders) {
			return localTypes.DIR_up, localTypes.Moving
		} else if requests_here(MyElev, MyOrders) || requests_below(MyElev, MyOrders) {
			return localTypes.DIR_down, localTypes.Moving
		} else {
			return localTypes.DIR_up, localTypes.Moving
		}
	case MyElev.Direction == localTypes.DIR_down:
		if requests_below(MyElev, MyOrders) {
			return localTypes.DIR_down, localTypes.Moving
		} else if requests_here(MyElev, MyOrders) || requests_above(MyElev, MyOrders) {
			return localTypes.DIR_up, localTypes.Moving
		} else {
			return localTypes.DIR_down, localTypes.Moving
		}
	case MyElev.Direction == localTypes.DIR_stop:
		if requests_here(MyElev, MyOrders) {
			return localTypes.DIR_stop, localTypes.Door_open
		} else if requests_above(MyElev, MyOrders) {
			return localTypes.DIR_up, localTypes.Moving
		} else if requests_below(MyElev, MyOrders) {
			return localTypes.DIR_down, localTypes.Moving
		} else {
			return localTypes.DIR_stop, localTypes.Idle
		}

	}
	return localTypes.DIR_stop, localTypes.Idle
}

func dir2Btntype(dir localTypes.MOTOR_DIR) localTypes.BUTTON_TYPE {
	if dir == localTypes.DIR_up {
		return localTypes.Button_hall_up
	} else if dir == localTypes.DIR_down {
		return localTypes.Button_hall_down
	} else if dir == localTypes.DIR_stop {
		panic("Invalid direction")
	}
	panic("No mototdir found???")
}

func requests_here(MyElev *localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	totalOrders := combineOrders(MyElev.CabCalls, MyOrders)
	for btn := 0; btn < localTypes.NUM_BUTTONS; btn++ {
		if totalOrders[MyElev.Floor][btn] {
			return true
		}
	}
	return false
}

func requests_above(MyElev *localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
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

func requests_below(MyElev *localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
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

func combineOrders(MyCabs [localTypes.NUM_FLOORS]bool, MyOrders localTypes.HMATRIX) [localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS]bool {
	var result [localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS]bool
	for i := 0; i < localTypes.NUM_FLOORS; i++ {
		result[i][0] = MyCabs[i]
		for j := 0; j < localTypes.NUM_BUTTONS; j++ {
			result[i][j] = MyOrders[i][j-1]
		}
	}
	return result
}

func addNewOrdersToLocal(newOrder localTypes.ORDER, MyOrders *localTypes.HMATRIX, MyElev localTypes.LOCAL_ELEVATOR_INFO) {
	for f := 0; f < localTypes.NUM_FLOORS; f++ {
		for btn := 0; btn < localTypes.NUM_BUTTONS-1; btn++ {
			(*MyOrders)[f][btn] = newOrder[MyElev.ElevID][f][btn]
		}
	}
}

func addNewOrdersToHMatrix(newOrder localTypes.ORDER, CombinedHMatrix *localTypes.HMATRIX) {
	for ID := range newOrder {
		for f := 0; f < localTypes.NUM_FLOORS; f++ {
			for btn := 0; btn < localTypes.NUM_BUTTONS-1; btn++ {
				if !CombinedHMatrix[f][btn] {
					(*CombinedHMatrix)[f][btn] = newOrder[ID][f][btn]
				}
			}
		}
	}
}
