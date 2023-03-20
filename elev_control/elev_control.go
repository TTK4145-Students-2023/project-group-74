package elev_control

import (
	"localTypes"
	"peers"
	"project-group-74/localTypes"
	"time"
)

//Channels:
//Output : To DLOCC : ElevInfoChan,NewHallRequestChan,FinishedHOrderChan
// To P2P   : TxP2PElevInfoChan
//Inputs : From DLOCC    : NewOrdersChan
// From Hardware : NewBtnPressChan, NewFloorChan
// From P2P      : RxP2pElevInfoChan

func RunElevator(
	TxElevInfoChan chan<- localTypes.LOCAL_ELEVATOR_INFO,
	RxElevInfoChan chan<- localTypes.LOCAL_ELEVATOR_INFO,
	TxNewHallRequestChan chan<- localTypes.BUTTON_INFO,
	RxNewHallRequestChan chan<- localTypes.BUTTON_INFO,
	TxFinishedHallOrderChan chan<- localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan chan<- localTypes.BUTTON_INFO,
	RxNewOrdersChan <-chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	TxP2PElevInfoChan chan<- localTypes.P2P_ELEV_INFO,
	RxP2PElevInfoChan <-chan localTypes.P2P_ELEV_INFO,
	NewFloorChan chan int,
	NewBtnPressChan <-chan localTypes.BUTTON_INFO) {

	redecideChan := make(chan bool)
	timeOutChan := make(chan bool)

	MyElev :=
		&localTypes.LOCAL_ELEVATOR_INFO{
			State:      localTypes.idle,
			Floor:      GetFloor(),
			Direction:  localTypes.MD_Stop,
			CabCalls:   [localTypes.NUM_FLOORS]bool,
			ElevatorID: network.MyIP,
		}
	LocalElevInitFloor(MyElev)
	MyOrders := &localTypes.HMATRIX        // FOR DRIVING combined with myCabCalls
	CombinedHMatrix := &localTypes.HMATRIX // FOR LIGHTS and reboot if you become master
	ForeignElevs := &[]localTypes.P2P_ELEV_INFO
	var timeOutTimer = time.Now()
	var bufferTimer *time.Timer = nil
	/*
		Initmaster:
			MHMatrixChan <- CombinedHMatrix //sends current CombinedHMatrix to itself
			MForeignElevChan <- ForeignElevs // sends current Foreignelevs to itself
	*/

	for {
		select {
		case newOrder := <-RxNewOrdersChan:
			AddNewOrders(newOrder, MyOrders, CombinedHMatrix, MyElev)
			UpdateOrderLights(MyElev, CombinedHMatrix)
			ForeignElevsChan <- ForeignElevs
			redecideChan <- true

		case newFloor := <-NewFloorChan:
			SetFloorIndicator(newFloor)
			if IsOrderAtFloor(newFloor) == 1 {
				PastElev := MyElev
				go ArrivedAtOrder(MyElev, newFloor) //Opendoors, wait, wait for them to press cab etc
				finishedOrder := GetFinOrder(newfloor, PastElev.MotorDirection)
				if finishedOrder.BUTTONTYPE == localTypes.BUTTON_CAB {
					RemoveOneOrderBtn(finishedOrder, MyElev)
					UpdateOrderLights(MyElev, CombinedHMatrix)
				} else {
					if IsMaster(MyElev.ElevatorID, peers.Peers) == true {
						RxFinishedHallOrderChan <- finishedOrder
					} else {
						TxFinishedHallOrderChan <- finishedOrder
					}

				}
			}
			MyElev.Floor = newFloor
			redecideChan <- true

		case newBtnPress := <-NewBtnPressChan:
			if newBtnPress.BUTTON_TYPE == localTypes.Button_Cab {
				AddOneNewOrderBtn(newBtnPress, MyElev)
				UpdateOrderLights(MyElev, CombinedHMatrix)
				redecideChan <- true
			} else {
				if !IsHOrderActive(newBtnPress, CombinedHMatrix) {
					if IsMaster(MyElev.ElevatorID, peers.Peers) == true {
						RxNewHallRequestChan <- newBtnPress
					} else {
						TxNewHallRequestChan <- newBtnPress
					}

				}
			}

		case <-redecideChan:
			if MyElev.State == localTypes.DoorOpen {
				bufferTimer = time.NewTimer(3 * time.Second)
				break
			} else if time.Since(bufferTimer.StartTime()) <= 3*time.Second {
				break
			} else if !bufferTimer.Stop() {
				bufferTimer.Stop()
				bufferTimer = nil
			}
			ChooseDirectionAndState(MyElev, MyOrders)
			if IsMaster(MyElev.ElevatorID, peers.Peers) == true {
				RxElevInfoChan <- MyElev
			} else {
				TxElevInfoChan <- MyElev
			}
			if MyElev.State == DoorOpen {
				NewFloorChan <- MyElev.Floor
			}

		case NewForeignInfo <- RxP2PElevInfoChan:
			ForeignElevs = NewForeignInfo
			AddLocalToForeignInfo(MyElev, ForeignElevs)
			go SendWithDelay(ForeignElevs, TxP2PElevInfoChan)

		default:
			if time.Since(timeOutTimer) >= 3*time.Second {
				if IsMaster(MyElev.ElevatorID, peers.Peers) == true {
					RxElevInfoChan <- MyElev
				} else {
					TxElevInfoChan <- MyElev
				}
				TxP2PElevInfoChan <- ForeignElevs
				timeOutTimer = time.Now()
			}
		}
	}

}
