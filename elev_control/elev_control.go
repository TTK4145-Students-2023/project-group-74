package elev_control

import "elev_control/elevio"


func RunElevator (chans ElevControlChannels){  
	
	MyElev:= ElevInfo{
		availability: Avaliable,
		lastfloor: GetFloor(),
		motorDirection : MD_Stop,
	}

	MyState ElevState = Idle

	for GetFloor() == -1 {
		SetMotorDirection(MD_Down)
	}
	SetMotorDirection(MD_Stop)
	MyElev.lastfloor=GetFloor()


	for{
		select{
		case newOrder := <- ElevControlChannels.NewOrders: 
			UpdateCurrentOrders(newOrder)
			UpdateHMatrix() /// ?????? 
			switch(MyState){
				case Idle:
					ElevInfo.MotorDirection = ChooseDirection(CurrentOrders,MyElev.lastfloor)
					SetMotorDirection(MyElev.MotorDirection)
					if MyElev.motorDirection == MD_Stop{
						ElevState=DoorOpen
						ArrivedAtOrder() // stop motor, change elevstate to dooropen, add to finished order. prepare for cab call(?)
					
					}
					else{
						ElevState=Moving
					}
				case DoorOpen:
					ArrivedAtOrder()
				case Moving:
					
					

			}
		case newFloor := <- ElevControlChannels.NewFloor:
			if newFloor==MyElev.lastfloor{
				break
			}
			MyElev.lastfloor=newFloor
			if OrderAtFloor(newfloor)==1{
				finishedOrder:=ArrivedAtOrder()
				ElevControlChannels.FinishedOrder <- finishedOrder
			}





		case finishedOrder := <- ElevControlChannels.FinishedOrder:

		case newBtnPress := <- ElevControlChannels.NewBtnpress:
		
			



		}
	}

}
	for {
		select {
		case a := <-drv_buttons:   // if btnpressed
			order_matrix[a.Button][a.Floor] = 1
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:  // if floorsensor has new value
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr: // if obstruction active
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop: // if stop active
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}



/////////////

input: 
next_order (from master or cab call)
Hall-call matrix (recived from master) 


internal:

action master call
	targetfloor=master_Call// maybe or maybe not
	drive
	while driving towards targetfloor 
		check if currentfloor && current_direction in hMatrix
			DoorOpen()
			output elev_output
			next_action -> drive_cab_call

	if current_floor=targetfloor	
		output elev_output
		next_action -> drive_cab_call

action drive_cab_call
	decide_targetfloor()
	while driving	
		check if currentfloor && current_direction in hMatrix
			stop, dooropenfunc
			output elev_output
			next_action -> drive_cab_call

		if current_floor=targetfloor	
			output elev_output
			next_action -> idle


action idle
	motor stop












































