package decision

import (
	"fmt"
	"project-group-74/localTypes"
	"project-group-74/decision/DLOCC"
	"project-group-74/network"
	"runtime"
)

func OrderAssigner(
	RxElevInfoChan <-chan localTypes.LOCAL_ELEVATOR_INFO,
	RxNewHallRequestChan <-chan localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan <-chan localTypes.BUTTON_INFO,
	TxNewOrdersChan chan<- map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	RxNewOrdersChan chan<- map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
) {

	ordersFromNetwork := make(chan localTypes.HRAInput)

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	go DLOCC.CombineHRAInput(
		RxElevInfoChan,
		RxNewHallRequestChan,
		RxFinishedHallOrderChan,
		ordersFromNetwork)

	/*mapOfElevators := make(map[string]types.HRAElevState) //Define a map with information on all elevators
	mapOfElevators[elevator.ID] = elevator

	orderActivatedChn   := make(chan types.OrderType)
	orderDeactivatedChn := make(chan types.OrderType)
	orderTimedOutChn    := make(chan types.OrderType)

	*/
	

	for {
		select {
		//case localID = <-OA.localIDch:
		//case assignerBehavior = <-orderAssignerBehaviorCh: // define this channel
		case newHRAInput := <-ordersFromNetwork:
			fmt.Printf("")
			if localTypes.IsMaster(network.MyIP, network.PeerList.Peers) {
				newOrders := DLOCC.ReassignOrders(newHRAInput, hraExecutable)

				TxNewOrdersChan <- newOrders
				RxNewOrdersChan <- newOrders
			}
		default:	

			/*switch IsMaster {
				case false: //This is info from the slaves, master vil send out
				case true:
					jsonBytes, err := json.Marshal(newHRAInput)
					if err != nil {
						fmt.Println("json.Marshal error: ", err)
						return
					}

					ret, err := exec.Command("../hall_request_assigner/"+ hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
					if err != nil {
						fmt.Println("exec.Command error: ", err)
						fmt.Println(string(ret))
						return
					}

					output := map[string][types.NUM_FLOORS][2]bool{}
					err = json.Unmarshal(ret, &output)
					if err != nil {
						fmt.Println("json.Unmarshal error: ", err)
						return
					}
					TxNewOrdersChan <- output
					RxNewOrdersChan <- output


					if localHallOrders, ok := output[localID]; ok {
						OA.localOrder <- localHallOrders
					}

					OA.ordersToSlave <- ret
				}
			case newOrdersInByteFormat := <-OA.ordersFromMaster:
				switch assignerBehavior {
				case OABehaviorMaster: //Do nothing since this is from master
				case OABehaviorSlave:
					output := map[string][types.NUM_FLOORS][2]bool{}
					err := json.Unmarshal(givenOrders, &output)
					if err != nil {
						fmt.Println("json.Unmarshal error: ", err)
						return
					}

					if localHallOrders, ok := output[localID]; ok {
						OA.localOrder <- localHallOrders
					}
				}*/
		}
	}
}
