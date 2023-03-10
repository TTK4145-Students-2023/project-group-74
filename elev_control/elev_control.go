package elev_control

import "elev_control/elevio"
import "types"




func RunElevator (chans ElevControlChannels){  
	
	MyElev:= &
	LOCAL_ELEVATOR_INFO{
		State: idle,
		Floor: GetFloor(),
		Direction : MD_Stop,
		Orders:[NUM_FLOORS][NUM_BUTTON]bool,
	}

	for GetFloor() == -1 {
		SetMotorDirection(MD_Down)
	}
	SetMotorDirection(MD_Stop)
	MyElev.Floor=GetFloor()


	for{
		select{
		case newOrder := <- ElevControlChannels.NewOrders: 
			AddNewOrdersToLocal(newOrder,MyElev)
			AddNewOrdersToForeign(newOrder,MyElev)
			redecideChan<-true
			
		case newFloor := <- ElevControlChannels.NewFloor:
			if newFloor==MyElev.lastfloor{
				break
			}
			MyElev.lastfloor=newFloor
			if OrderAtFloor(newfloor)==1{
				ArrivedAtOrder() //Opendoors, wait, wait for them to press cab etc
				finishedOrder:=GetOrder(newfloor,MyElev.MotorDirection)
				ElevControlChannels.FinishedOrder <- finishedOrder
			}




		case finishedOrder := <- ElevControlChannels.FinishedOrder:
			registerFinishedOrder() //check if cab call or H call,
			ExternalFinishedOrderChan <- finishedOrder
			redecideChan <- true

			

		case newBtnPress := <- ElevControlChannels.NewBtnpress:
			if newBtnPress.ButtonType==BT_Cab{
				
			}
			

		case redecide: <- redecideChan
			ChooseDirectionAndState(MyElev)
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












































