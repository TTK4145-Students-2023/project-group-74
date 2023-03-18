package elev_control

func ChooseDirectionAndState(MyElev *LOCAL_ELEVATOR_INFO, MyOrders HMATRIX) {
	newDirAndState[2]=findDirection(MyElev)
	SetMotorDirection(newDirAndState[0])
	MyElev.MotorDirection=newDirAndState[0]
	MyElev.State=newDirAndState[1]	
}

func findDirection(MyElev LOCAL_ELEVATOR_INFO, MyOrders HMATRIX) [2]int {
	switch{
	case MyElev.MotorDirection==MD_Up:
		if requests_above(MyElev,CurrentHmatrix) {
			return [2]int{MD_Up, Moving}
		} else if requests_here(MyElev,CurrentHmatrix)||requests_below(MyElev,CurrentHmatrix) {
			return [2]int{MD_Down, Moving}
		} else {
			return [2]int{MD_Up, Moving}	
		}
	case MyElev.MotorDirection==MD_Down:
		if requests_below(MyElev,CurrentHmatrix) {
			return [2]int{MD_Down, Moving}
		} else if requests_here(MyElev,CurrentHmatrix)||requests_above(MyElev,CurrentHmatrix) {
			return [2]int{MD_Up, Moving}
		}  else {
			return [2]int{MD_Down, Moving}	
		}
	case MyElev.MotorDirection==MD_Stop:
		if requests_here(MyElev,CurrentHmatrix) {
			return [2]int{MD_Stop, DoorOpen}
		} else if requests_above(MyElev,CurrentHmatrix) {
			return [2]int{MD_Up, Moving}
		} else if requests_below(MyElev,CurrentHmatrix) {
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

func removeOneOrderBtn(finishedOrder BUTTON_INFO, MyElev *LOCAL_ELEV_INFO){
	MyElev.Orders[finishedOrder.floor][finishedOrder.button]=0
}

func IsHOrderActive(newOrder BUTTON_INFO, CurrentHMatrix HMATRIX) bool{ //neccecary?
	if CurrentHMatrix[newOrder.floor][newOrder.button]==0{
		return false
	}
	return true
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

func localElevInitFloor(MyElev *LOCAL_ELEVATOR_INFO){
	for GetFloor() == -1 {
		SetMotorDirection(MD_Down)
	}
		SetMotorDirection(MD_Stop)
	MyElev.Floor=GetFloor()
}

func requests_here(myElev LOCAL_ELEV_INFO,MyOrders HMATRIX)bool{
	totalOrders:= append(MyElev.CabCalls[:],MyOrders)
	for btn=0;btn<NUM_BUTTON;btn++{
		if totalOrders[myElev.Floor][btn]{
			return true
		}
	}
	return false
}

func requests_above(myElev LOCAL_ELEV_INFO,MyOrders HMATRIX)bool{
	totalOrders:= append(MyElev.CabCalls[:],MyOrders)
	for f:=myElev.Floor+1;f<NUM_FLOORS;f++{
		for btn=0;btn<NUM_BUTTON;btn++{
			if totalOrders[f][btn]{
				return true
			}
		}
	}
	return false
}

func requests_below(myElev LOCAL_ELEV_INFO,MyOrders HMATRIX)bool{
	totalOrders:= append(MyElev.CabCalls[:],MyOrders)
	for f:=0;f<myElev.Floor;f++{
		for btn=0;btn<NUM_BUTTON;btn++{
			if totalOrders[f][btn]{
				return true
			}
		}
	}
	return false
}


func AddNewOrders(newOrder ORDER, MyOrders *HMATRIX,CombinedHMatrix *HMATRIX){
	addNewOrdersToLocal(newOrder,MyOrders)
	addNewOrdersToHMatrix(newOrder,CombinedHMatrix)
}

func addNewOrdersToLocal(newOrder ORDER, MyOrders *HMATRIX){
	for f:=0;f<NUM_FLOORS;f++{
		for btn:=0;btn<NUM_BUTTON-1;btn++{
			MyOrders[f][btn] = neworder[MyElev.ElevID][f][btn]
		}
	}
}

func addNewOrdersToHMatrix(newOrder ORDER,CombinedHMatrix *HMATRIX){
	ElevIDs :=make([]string, 0, len(newOrder))
	for ID:=range newOrder{
		ElevIDs=append(ElevIDs,ID)
	}
	for _,ID:= range ElevIDs{
		for f:=0;f<NUM_FLOORS;f++{
			for btn:=0;btn<NUM_BUTTON-1;btn++{
				if CombinedHMatrix[f][btn] ==0{
					CombinedHMatrix[f][btn]=newOrder[ID][f][btn]
				}
			}
		}	 
	}
}


func AddLocalToForeignInfo(MyElev LOCAL_ELEVATOR_INFO, ForeignElevs *P2P_ELEV_INFO){
	for ForeignElev:=0;ForeignElev<len(ForeignElevs);ForeignElev++{
		if ForeignElevs[ForeignElev].ElevID==MyElev.ElevID{
			ForeignElevs[ForeignElev]=MyElev
		}
	}
}

func updateLights(MyElev LOCAL_ELEV_INFO, CurrentHMatrix HMATRIX){
	for f:=0;f<NUM_FLOORS;f++{
		SetButtonLamp(BUTTON_CAB,f,MyElev.CabCalls[f])
		for btn:=0;btn<NUM_BUTTONS-1;btn++{
			SetButtonLamp(btn,f,CurrentHMatrix[f][btn])
		}
	}
}

//undefined funcs

/*  
	if foreignelevs.ElevID in newOrder.key
		update the current foreignorder. // Only needs to update foreignelevs with the neworders if neworders should contain MyOrders as well. 
*/	


func ArrivedAtOrder(
	MyElev *LOCAL_ELEVATOR_INFO,
	newFloor int,
	TxElevInfoChan chan<- LOCAL_ELEVATOR_INFO,
	RxElevInfoChan chan<- LOCAL_ELEVATOR_INFO){
	
	Myelev.Direction=MD_Stop
	MyElev.Floor=newfloor
	MyElev.State=DoorOpen
	SetMotorDirection(MD_Stop)

	if IsMaster(MyElev.ElevatorID,peers.Peers)==true{
		RxElevInfoChan <- MyElev
	}else TxElevInfoChan <- MyElev

	SetDoorOpenLamp(true)
	doorTimer=time.NewTimer(3 * time.Second)
	for time.Since(doorTimer.StartTime()) <= 3 * time.Second{
		Sleep(100*time.Millisecond)
	}
	doorTimer.Stop()
	SetDoorOpenLamp(false)
	MyElev.State=Idle
}


func registerFinishedCabOrder(finishedOrder BUTTON_INFO, MyElev *LOCAL_ELEV_INFO){
}

func registerFinishedHallOrder(finishedOrder BUTTON_INFO, MyElev *LOCAL_ELEV_INFO){

}