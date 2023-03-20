package elevio

import (
	"project-group-74/elev_control/elevio"
	"project-group-74/localTypes"
	"time"
)

// used as goroutine

func ArrivedAtOrder(
	MyElev *localTypes.LOCAL_ELEVATOR_INFO) {

	*&(MyElev).Direction = localTypes.DIR_stop
	MyElev.Floor = elevio.GetFloor
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
	timer := time.NewTimer(localTypes.P2P_UPDATE_INTERVAL * time.Second)
	<-timer.C
	TxChannel <- foreignElevs
}

// Private funcs
func ChooseDirectionAndState(MyElev *localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) {
	newDirAndState[2] = findDirection(MyElev)
	SetMotorDirection(newDirAndState[0])
	MyElev.MotorDirection = newDirAndState[0]
	MyElev.State = newDirAndState[1]
}

func IsHOrderActive(newOrder localTypes.BUTTON_INFO, CurrentHMatrix localTypes.HMATRIX) bool { //neccecary?
	if CurrentHMatrix[newOrder.Floor][newOrder.button] == 0 {
		return false
	}
	return true
}

func IsOrderAtFloor(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	btntype := dir2Btntype(MyElev.Direction)
	if MyElev.CabCalls[GetFloor()] || MyOrders[GetFloor()][btntype] == 1 {
		return true
	}
	return false
}

func AddNewOrders(newOrder localTypes.ORDER, MyOrders *localTypes.HMATRIX, CombinedHMatrix *localTypes.HMATRIX) {
	addNewOrdersToLocal(newOrder, MyOrders)
	addNewOrdersToHMatrix(newOrder, CombinedHMatrix)
}

func AddLocalToForeignInfo(MyElev localTypes.LOCAL_ELEVATOR_INFO, ForeignElevs *localTypes.P2P_ELEV_INFO) {
	for ForeignElev := 0; ForeignElev < len(ForeignElevs); ForeignElev++ {
		if ForeignElevs[ForeignElev].ElevID == MyElev.ElevID {
			ForeignElevs[ForeignElev] = MyElev
		}
	}
}

func UpdateOrderLights(MyElev localTypes.LOCAL_ELEVATOR_INFO, CurrentHMatrix localTypes.HMATRIX) {
	for f := 0; f < localTypes.NUM_FLOORS; f++ {
		SetButtonLamp(localTypes.Button_Cab, f, MyElev.CabCalls[f])
		for btn := 0; btn < localTypes.NUM_BUTTONS-1; btn++ {
			SetButtonLamp(btn, f, CurrentHMatrix[f][btn])
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
	if MyElev.CabCalls[newOrder.Floor] == false {
		MyElev.CabCalls[newOrder.Floor] = true
	}
}

//Internal funcs

func findDirection(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) [2]int {
	switch {
	case MyElev.Direction == localTypes.DIR_up:
		if requests_above(MyElev, MyOrders) {
			return [2]int{localTypes.DIR_up, localTypes.Moving}
		} else if requests_here(MyElev, MyOrders) || requests_below(MyElev, MyOrders) {
			return [2]int{localTypes.DIR_down, localTypes.Moving}
		} else {
			return [2]int{localTypes.DIR_up, localTypes.Moving}
		}
	case MyElev.Direction == localTypes.DIR_down:
		if requests_below(MyElev, MyOrders) {
			return [2]int{localTypes.DIR_down, localTypes.Moving}
		} else if requests_here(MyElev, MyOrders) || requests_above(MyElev, MyOrders) {
			return [2]int{localTypes.DIR_up, localTypes.Moving}
		} else {
			return [2]int{localTypes.DIR_down, localTypes.Moving}
		}
	case MyElev.Direction == localTypes.DIR_stop:
		if requests_here(MyElev, MyOrders) {
			return [2]int{localTypes.DIR_stop, localTypes.Door_open}
		} else if requests_above(MyElev, MyOrders) {
			return [2]int{localTypes.DIR_up, localTypes.Moving}
		} else if requests_below(MyElev, MyOrders) {
			return [2]int{localTypes.DIR_down, localTypes.Moving}
		} else {
			return [2]int{localTypes.DIR_stop, localTypes.Idle}
		}

	}
}

func dir2Btntype(dir localTypes.MOTOR_DIR) localTypes.BUTTON_TYPE {
	if dir == localTypes.DIR_up {
		return localTypes.Button_hall_up
	} else if dir == localTypes.DIR_down {
		return localTypes.Button_hall_down
	} else if dir == localTypes.DIR_stop {
		panic("Invalid direction")
	}
}

func requests_here(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	MyCabsT := transposeMyCabCalls(MyElev.CabCalls)
	totalOrders := append(MyCabsT[:], MyOrders[:])
	for btn := 0; btn < localTypes.NUM_BUTTONS; btn++ {
		if totalOrders[MyElev.Floor][btn] {
			return true
		}
	}
	return false
}

func requests_above(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	MyCabsT := transposeMyCabCalls(MyElev.CabCalls)
	totalOrders := append(MyCabsT[:], MyOrders[:])
	for f := MyElev.Floor + 1; f < localTypes.NUM_FLOORS; f++ {
		for btn := 0; btn < localTypes.NUM_BUTTONS; btn++ {
			if totalOrders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(MyElev localTypes.LOCAL_ELEVATOR_INFO, MyOrders localTypes.HMATRIX) bool {
	MyCabsT := transposeMyCabCalls(MyElev.CabCalls)
	totalOrders := append(MyCabsT[:], MyOrders[:])
	for f := 0; f < MyElev.Floor; f++ {
		for btn := 0; btn < localTypes.NUM_BUTTONS; btn++ {
			if totalOrders[f][btn] {
				return true
			}
		}
	}
	return false
}

func transposeMyCabCalls(CabCalls [localTypes.NUM_FLOORS]bool) [][]bool {
	MyCabsT := make([][]bool, len(CabCalls))
	for i := range MyCabsT {
		MyCabsT[i] = []bool{CabCalls[i]}
	}
	return MyCabsT
}

func addNewOrdersToLocal(newOrder localTypes.ORDER, MyOrders *localTypes.HMATRIX) {
	for f := 0; f < localTypes.NUM_FLOORS; f++ {
		for btn := 0; btn < localTypes.NUM_BUTTON-1; btn++ {
			(*MyOrders)[f][btn] = neworder[MyElev.ElevID][f][btn]
		}
	}
}

func addNewOrdersToHMatrix(newOrder localTypes.ORDER, CombinedHMatrix *localTypes.HMATRIX) {
	for ID := range newOrder {
		for f := 0; f < localTypes.NUM_FLOORS; f++ {
			for btn := 0; btn < localTypes.NUM_BUTTON-1; btn++ {
				if CombinedHMatrix[f][btn] == 0 {
					(*CombinedHMatrix)[f][btn] = newOrder[ID][f][btn]
				}
			}
		}
	}
}
