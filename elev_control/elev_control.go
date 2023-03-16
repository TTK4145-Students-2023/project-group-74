package elev_control
jj
import "elev_control/elevio"
import "project-group-74/network/subs/localip"
import "localTypes"


//Channels:
	//Output : To DLOCC : ElevInfoChan,NewHallRequestChan,FinishedHOrderChan
			// To P2P   : TxP2PElevInfoChan
	//Inputs : From DLOCC    : NewOrdersChan
			// From Hardware : NewBtnPressChan, NewFloorChan
			// From P2P      : RxP2pElevInfoChan

func RunElevator (ElevInfoChan chan,NewHallRequestChan chan,FinishedHOrderChan chan, 
				  TxP2PElevInfoChan chan, NewOrdersChan chan, NewBtnPressChan chan,
				  NewFloorChan chan, RxP2pElevInfoChan chan){  
	
	MyElev:=
	&LOCAL_ELEVATOR_INFO{
		State: idle,
		Floor: GetFloor(),
		Direction : MD_Stop,
		Orders:[NUM_FLOORS][2]bool,
		ElevatorID :network.localip.LocalIP(),
	}
	CurrentHMatrix := &HMATRIX
	localElevInit(MyElev,CurrentHMatrix)
	ForeignElevs := []FOREIGN_ELEVATOR_INFO

	for{
		select{
		case newOrder := <- ElevControlChannels.NewOrders: 
			AddNewOrdersToLocal(newOrder,MyElev,CurrentHMatrix)
			UpdateLights(MyElev,CurrentHMatrix)
			ElevInfoChan <- MyElev
			AddNewOrdersToForeign(newOrder,ForeignElevs)
			ForeignElevsChan<-ForeignElevs
			redecideChan<-true
			
		case newFloor := <- ElevControlChannels.NewFloorChan:
			if IsOrderAtFloor(newFloor)==1{
				ArrivedAtOrder(MyElev) //Opendoors, wait, wait for them to press cab etc
				finishedOrder:=GetOrder(newfloor,MyElev.MotorDirection)
				if finishedOrder.BUTTONTYPE==BUTTON_CAB{
					registerFinishedCabOrder(finishedOrder,MyElev)
					UpdateLights(MyElev,CurrentHMatrix)
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
				UpdateLights(MyElev,CurrentHMatrix)
				ElevInfoChan <- MyElev
				ElevControlChannels.redecideChan<-true
			} else{
				if !IsHOrderActive(newBtnPress,CurrentHMatrix){
					newHcallRequestChan <- newBtnPress
				}
			}

		case ShouldRedecide:= <- ElevControlChannels.redecideChan:
			ChooseDirectionAndState(MyElev, CurrentHMatrix)
			ElevInfoChan <- MyElev
			if MyElev.State ==DoorOpen{
				ElevControlChannels.NewFloorChan <- MyElev.Floor
			}

		case timeOut := <- TimeOutChan:
			ElevInfoChan <- MyElev
			ForeignElevsChan <- ForeignElevs

		}
	}

}


































