package elev_control

func ChooseDirectionAndState(MyElev *LOCAL_ELEVATOR_INFO) {
	newDirAndState[2]=findDirection(MyElev)
	SetMotorDirection(newDirAndState[0])
	MyElev.MotorDirection=newDirAndState[0]
	MyElev.State=newDirAndState[1]	
}

func findDirection(MyElev LOCAL_ELEVATOR_INFO) [2]int {
	switch{
	case MyElev.MotorDirection==MD_Up:
		if requests_above(MyElev) {
			return [2]int{MD_Up, Moving}
		} else if requests_here(MyElev)||requests_below(MyElev) {
			return [2]int{MD_Down, Moving}
		} else {
			return [2]int{MD_Up, Moving}	
		}
	case MyElev.MotorDirection==MD_Down:
		if requests_below(MyElev) {
			return [2]int{MD_Down, Moving}
		} else if requests_here(MyElev)||requests_above(MyElev) {
			return [2]int{MD_Up, Moving}
		}  else {
			return [2]int{MD_Down, Moving}	
		}
	case MyElev.MotorDirection==MD_Stop:
		if requests_here(MyElev) {
			return [2]int{MD_Stop, DoorOpen}
		} else if requests_above(MyElev) {
			return [2]int{MD_Up, Moving}
		} else if requests_below(MyElev) {
			return [2]int{MD_Down, Moving}
		} else {
			return [2]int{MD_Stop, Idle}		
		}		
	
	}
}

func addOneNewOrderBtn(newOrder BUTTON_INFO, MyElev *LOCAL_ELEVATOR_INFO){ //neccecary?
	if MyElev.Orders[newOrder.floor][newOrder.button]==0{
		MyElev.Orders[newOrder.floor][newOrder.button]=1
	}
}

func IsOrderAtFloor(MyElev LOCAL_ELEVATOR_INFO) bool {
	btntype := dir2Btntype(MyElev.motordirection)
	if MyElev.Orders[GetFloor()][btntype] == 1 {
		return true
	}
	return false
}

func dir2Btntype(dir motordirection) ButtonType {
	if dir == MD_Up {
		return BT_HallUp
	} else if dir == MD_Down {
		return BT_HallDown
	} else if dir == MD_Stop {
		panic("Invalid direction")
	}
}

func localElevInit(MyElev *LOCAL_ELEVATOR_INFO){
	for GetFloor() == -1 {
		SetMotorDirection(MD_Down)
	}
		SetMotorDirection(MD_Stop)
	MyElev.Floor=GetFloor()
}


Undefinedfuncs: 

func requests_here(myElev LOCAL_ELEV_INFO)bool{

}

func requests_above(myElev LOCAL_ELEV_INFO)bool{

}

func requests_below(myElev LOCAL_ELEV_INFO)bool{

}

func ArrivedAtOrder(finOrderChan chan ButtonEvent, MyElev *LOCAL_ELEVATOR_INFO){

}

func UpdateLights(MyElev LOCAL_ELEV_INFO){
	Oppdater lys når ny local elev info er tilgjengelig. 
	Burde gjøres hver gang Local Elevinfo.ORders
	blir oppdatert. 
}

func registerFinishedCabOrder(finishedOrder BUTTON_INFO, MyElev *LOCAL_ELEV_INFO){
}

func registerFinishedHallOrder(finishedOrder BUTTON_INFO, MyElev *LOCAL_ELEV_INFO){

}