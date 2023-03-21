package elev_control

import (
	"fmt"
	"project-group-74/elev_control/elevio"
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

	MyElev :=
		localTypes.LOCAL_ELEVATOR_INFO{
			State:     localTypes.Idle,
			Floor:     elevio.GetFloor(),
			Direction: localTypes.DIR_stop,
			CabCalls:  [localTypes.NUM_FLOORS]bool{},
			ElevID:    localTypes.MyIP,
		}

	MyElevPtr := &MyElev
	elevio.LocalElevInitFloor(MyElevPtr)
	var MyOrders localTypes.HMATRIX        // FOR DRIVING combined with myCabCalls
	var CombinedHMatrix localTypes.HMATRIX // FOR LIGHTS and reboot if you become master
	ForeignElevs := make(localTypes.P2P_ELEV_INFO, 0)
	ForeignElevsPtr := &ForeignElevs
	//var timeOutTimer = time.Now()
	p2pTicker := time.NewTicker(localTypes.P2P_UPDATE_INTERVAL * time.Millisecond)
	fmt.Printf("  running elev:    %q\n", MyElev)
	for {
		select {
		case newOrder := <-RxNewOrdersChan:
			elevio.AddNewOrders(newOrder, &MyOrders, &CombinedHMatrix, MyElev)
			elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
			TxP2PElevInfoChan <- ForeignElevs
			redecideChan <- true

		case newFloor := <-NewFloorChan:
			elevio.SetFloorIndicator(newFloor)
			if elevio.IsOrderAtFloor(MyElev, MyOrders) == true {
				finishedOrder := elevio.GetFinOrder(newFloor, MyElev.Direction)
				if finishedOrder.Button == localTypes.Button_Cab {
					elevio.RemoveOneOrderBtn(finishedOrder, MyElevPtr)
					elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				} else {
					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) == true {
						RxFinishedHallOrderChan <- finishedOrder
					} else {
						TxFinishedHallOrderChan <- finishedOrder
					}
				}
				elevio.ArrivedAtOrder(MyElevPtr) //Opendoors, wait, wait for them to press cab etc
			}
			MyElev.Floor = newFloor
			redecideChan <- true

		case newBtnPress := <-NewBtnPressChan:
			fmt.Printf("  Newbtnpress    %q\n", newBtnPress)
			if newBtnPress.Button == localTypes.Button_Cab {
				elevio.AddOneNewOrderBtn(newBtnPress, MyElevPtr)
				elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				redecideChan <- true
			} else {
				if !elevio.IsHOrderActive(newBtnPress, CombinedHMatrix) {
					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) == true {
						RxNewHallRequestChan <- newBtnPress
					} else {
						TxNewHallRequestChan <- newBtnPress
					}

				}
			}

		case <-redecideChan:
			if MyElev.State == localTypes.Door_open {
				break
			}
			elevio.ChooseDirectionAndState(MyElevPtr, MyOrders)
			if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) == true {
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

		case <-p2pTicker.C:
			elevio.AddLocalToForeignInfo(MyElev, ForeignElevsPtr)
			fmt.Printf("  sending foreignelevs:    %q\n", ForeignElevs)
			TxP2PElevInfoChan <- ForeignElevs

			/*default:
			fmt.Printf("  defaul in elev_control   \n")
			if time.Since(timeOutTimer) >= 3*time.Second {
				if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) == true {
					RxElevInfoChan <- MyElev
				} else {
					TxElevInfoChan <- MyElev
				}
				TxP2PElevInfoChan <- ForeignElevs
				timeOutTimer = time.Now()
			}*/
		}
	}

}
