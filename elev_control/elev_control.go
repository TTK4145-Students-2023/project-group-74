package elev_control
 jabba dabba
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

	localElevInit(MyElev)

	for{
		select{
		case newOrder := <- ElevControlChannels.NewOrders: 
			AddNewOrdersToLocal(newOrder,MyElev)
			UpdateLights(MyElev)
			ElevInfoChan <- MyElev
			AddNewOrdersToForeign(newOrder,MyElev)
			redecideChan<-true
			
		case newFloor := <- ElevControlChannels.NewFloorChan:
			if IsOrderAtFloor(newFloor)==1{
				ArrivedAtOrder() //Opendoors, wait, wait for them to press cab etc
				finishedOrder:=GetOrder(newfloor,MyElev.MotorDirection)
				if finishedOrder.BUTTONTYPE==BUTTON_CAB{
					registerFinishedCabOrder(finishedOrder,MyElev)
					UpdateLights(MyElev)
					ElevInfoChan <- MyElev
				} else {
					registerFinishedHallOrder(finishedOrder,MyElev)
					ElevControlChannels.FinishedOrderChan <- finishedOrder
				}
				MyElev.Floor=newFloor

				redecideChan <- true
			}
			MyElev.Floor=newFloor
			UpdateLights(MyElev)
			ElevInfoChan <- MyElev

		case newBtnPress := <- ElevControlChannels.NewBtnpressChan:
			if newBtnPress.BUTTON_TYPE==Button_Cab{
				addOneNewOrderBtn(newBtnPress,MyElev)
				UpdateLights(MyElev)
				ElevInfoChan <- MyElev
				ElevControlChannels.redecideChan<-true
			} else{
				newHcallRequestChan <- newBtnPress
			}

		case ShouldRedecide:= <- ElevControlChannels.redecideChan:
			ChooseDirectionAndState(MyElev)
			ElevInfoChan <- MyElev
			if MyElev.State ==DoorOpen{
				ElevControlChannels.NewFloorChan <- MyElev.Floor
			}
		}
	}

}


































