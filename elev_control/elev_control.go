package elev_control
import "elev_control/elevio"
import "project-group-74/network/subs/localip"
import "localTypes"
import "time"


//Channels:
	//Output : To DLOCC : ElevInfoChan,NewHallRequestChan,FinishedHOrderChan
			// To P2P   : TxP2PElevInfoChan
	//Inputs : From DLOCC    : NewOrdersChan
			// From Hardware : NewBtnPressChan, NewFloorChan
			// From P2P      : RxP2pElevInfoChan

func RunElevator(
			TxElevInfoChan chan<- LOCAL_ELEVATOR_INFO,
			RxElevInfoChan chan<- LOCAL_ELEVATOR_INFO,
			TxNewHallRequestChan chan<- BUTTON_INFO,
			RxNewHallRequestChan chan<- BUTTON_INFO,
			TxFinishedHallOrderChan chan<- BUTTON_INFO,
			RxFinishedHallOrderChan chan<- BUTTON_INFO,
			RxNewOrdersChan <-chan map[string][types.NUM_FLOORS][types.NUM_BUTTONS-1]bool,
			TxP2PElevInfoChan chan<- P2P_ELEV_INFO,
			RxP2PElevInfoChan <-chan P2P_ELEV_INFO,
			NewFloorChan chan int,
			NewBtnPressChan <-chan BUTTON_INFO){

	redecideChan := make(chan bool)
	timeOutChan := make(chan bool)

	MyElev:=
	&LOCAL_ELEVATOR_INFO{
		State: idle,
		Floor: GetFloor(),
		Direction : MD_Stop,
		CabCalls:[NUM_FLOORS]bool,
		ElevatorID :network.MyIP,
	}
	localElevInitFloor(MyElev)
	MyOrders := &HMATRIX // FOR DRIVING combined with myCabCalls 
	CombinedHMatrix := &HMATRIX // FOR LIGHTS and reboot if you become master 
	ForeignElevs := &[]FOREIGN_ELEVATOR_INFO
	var timeOutTimer *time.Timer
	var bufferTimer *time.Timer = nil
/*
	Initmaster: 
		MHMatrixChan <- CombinedHMatrix //sends current CombinedHMatrix to itself
		MForeignElevChan <- ForeignElevs // sends current Foreignelevs to itself
*/	
	
	for{
		select{
		case newOrder := <- RxNewOrdersChan: 
			AddNewOrders(newOrder,MyOrders,CombinedHMatrix)
			UpdateLights(MyElev,CombinedHMatrix)
			ForeignElevsChan<-ForeignElevs
			redecideChan<-true
			
		case newFloor := <- NewFloorChan:
			SetFloorIndicator(newFloor)
			if IsOrderAtFloor(newFloor)==1{
				PastElev:=MyElev
				go ArrivedAtOrder(MyElev,newFloor) //Opendoors, wait, wait for them to press cab etc
				finishedOrder:=GetOrder(newfloor,PastElev.MotorDirection)
				if finishedOrder.BUTTONTYPE==BUTTON_CAB{
					removeOneOrderBtn(finishedOrder,MyElev)
					UpdateLights(MyElev,CombinedHMatrix)
				} else {
					ElevControlChannels.FinishedOrderChan <- finishedOrder
				}				
			}
			MyElev.Floor=newFloor
			redecideChan <- true

		case newBtnPress := <- NewBtnPressChan:
			if newBtnPress.BUTTON_TYPE==Button_Cab{
				addOneNewOrderBtn(newBtnPress,MyElev)
				UpdateLights(MyElev,CombinedHMatrix)
				redecideChan<-true
			} else{
				if !IsHOrderActive(newBtnPress,CombinedHMatrix){
					TxNewHallRequestChan <- newBtnPress
				}
			}

		case ShouldRedecide := <- redecideChan:
			if MyElev.State == DoorOpen{
				bufferTimer = time.NewTimer(3 * time.Second)
				break
			}else if time.Since(bufferTimer.StartTime()) <= 3 * time.Second{
				break
			}else if !bufferTimer.Stop(){
				bufferTimer.Stop()
				bufferTimer=nil
			}
			ChooseDirectionAndState(MyElev, CombinedHMatrix)
			if IsMaster(MyElev.ElevatorID,peers.Peers)==true{
				RxElevInfoChan <- MyElev
			}else TxElevInfoChan <- MyElev
			if MyElev.State == DoorOpen{
				NewFloorChan <- MyElev.Floor
			}
		
		case NewForeignInfo <- RxP2PElevInfoChan:
			ForeignElevs=NewForeignInfo
			AddLocalToForeignInfo(MyElev,ForeignElevs)
			TxP2PElevInfoChan<-ForeignElevs

		case timeOut := <- TimeOutChan:
			if IsMaster(MyElev.ElevatorID,peers.Peers)==true{
				RxElevInfoChan <- MyElev
			}else {TxElevInfoChan <- MyElev}
			TxP2PElevInfoChan <- ForeignElevs

			
		default:
			if timeOutTimer==nil{
				timeOutTimer=time.NewTimer(3 * time.Second)
			}else if time.Since(timeOutTimer.StartTime()) >= 3 * time.Second{
				timeOutChan<-true
				timeOutTimer.Stop()
				timeOutTimer=nil
			}
		}
	}

}


































