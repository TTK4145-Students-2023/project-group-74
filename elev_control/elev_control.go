package elev_control

import (
	"project-group-74/elev_control/elevio"
	"project-group-74/localTypes"
	"project-group-74/network"
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

	MyElev :=
		localTypes.LOCAL_ELEVATOR_INFO{
			State:     localTypes.Idle,
			Floor:     elevio.GetFloor(),
			Direction: localTypes.DIR_stop,
			CabCalls:  [localTypes.NUM_FLOORS]bool{},
			ElevID:    network.MyIP,
		}

	MyElevPtr := &MyElev
	elevio.LocalElevInitFloor(MyElevPtr)
	var MyOrders localTypes.HMATRIX        // FOR DRIVING combined with myCabCalls
	var CombinedHMatrix localTypes.HMATRIX // FOR LIGHTS and reboot if you become master
	ForeignElevs := make(localTypes.P2P_ELEV_INFO, 0)
	ForeignElevsPtr := &ForeignElevs
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
			elevio.AddNewOrders(newOrder, &MyOrders, &CombinedHMatrix)
			elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
			TxP2PElevInfoChan <- ForeignElevs
			redecideChan <- true

		case newFloor := <-NewFloorChan:
			elevio.SetFloorIndicator(newFloor)
			if elevio.IsOrderAtFloor(MyElev, MyOrders) == true {
				PastElev := MyElev
				go elevio.ArrivedAtOrder(MyElevPtr) //Opendoors, wait, wait for them to press cab etc
				finishedOrder := elevio.GetFinOrder(newfloor, PastElev.Direction)
				if finishedOrder.Button == localTypes.Button_Cab {
					elevio.RemoveOneOrderBtn(finishedOrder, MyElevPtr)
					elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				} else {
					if localTypes.IsMaster(MyElev.ElevID, network.PeerList.Peers) == true {
						RxFinishedHallOrderChan <- finishedOrder
					} else {
						TxFinishedHallOrderChan <- finishedOrder
					}

				}
			}
			MyElev.Floor = newFloor
			redecideChan <- true

		case newBtnPress := <-NewBtnPressChan:
			if newBtnPress.Button == localTypes.Button_Cab {
				elevio.AddOneNewOrderBtn(newBtnPress, MyElevPtr)
				elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				redecideChan <- true
			} else {
				if !elevio.IsHOrderActive(newBtnPress, CombinedHMatrix) {
					if localTypes.IsMaster(MyElev.ElevID, network.PeerList.Peers) == true {
						RxNewHallRequestChan <- newBtnPress
					} else {
						TxNewHallRequestChan <- newBtnPress
					}

				}
			}

		case <-redecideChan:
			if MyElev.State == localTypes.Door_open {
				bufferTimer = time.NewTimer(3 * time.Second)
				break
			} else if time.Since(bufferTimer.StartTime()) <= 3*time.Second {
				break
			} else if !bufferTimer.Stop() {
				bufferTimer.Stop()
				bufferTimer = nil
			}
			elevio.ChooseDirectionAndState(MyElevPtr, MyOrders)
			if localTypes.IsMaster(MyElev.ElevID, network.PeerList.Peers) == true {
				RxElevInfoChan <- MyElev
			} else {
				TxElevInfoChan <- MyElev
			}
			if MyElev.State == localTypes.Door_open {
				NewFloorChan <- MyElev.Floor
			}

		case NewForeignInfo := <-RxP2PElevInfoChan:
			ForeignElevs = NewForeignInfo
			elevio.AddLocalToForeignInfo(MyElev, ForeignElevsPtr)
			go elevio.SendWithDelay(ForeignElevs, TxP2PElevInfoChan)

		default:
			if time.Since(timeOutTimer) >= 3*time.Second {
				if localTypes.IsMaster(MyElev.ElevID, network.PeerList.Peers) == true {
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
