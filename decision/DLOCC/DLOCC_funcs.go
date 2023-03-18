package decision

func orderWatchdog(
	orderActivatedChn   <-chan   types.OrderType, 
	orderDeactivatedChn <-chan   types.OrderType, 
	orderTimedOutChn      chan<- types.OrderType) {

	var orderTimeouts [types.NUM_FLOORS][types.NUM_BUTTONS]time.Time
	var zeroTime             = time.Time{}
	pollOrderTimeoutsTicker := time.NewTicker(ORDER_WATCHDOG_POLL_RATE)
	
	for {
		select {
		case order := <-orderActivatedChn:
			timeout := orderTimeouts[order.Floor][order.Button]
			if timeout.IsZero() {
				orderTimeouts[order.Floor][order.Button] = time.Now().Add(ORDER_TIMEOUT_PERIOD)
			}

		case order := <-orderDeactivatedChn:
			orderTimeouts[order.Floor][order.Button] = zeroTime

		case <-pollOrderTimeoutsTicker.C:
			for floor := 0; floor < types.NUM_FLOORS; floor++ {
				for button := 0; button < types.NUM_BUTTONS; button++ {
					timeout := orderTimeouts[floor][button]

					if !timeout.IsZero() && timeout.Before(time.Now()) {
						orderTimeouts[floor][button] = zeroTime
						var order types.OrderType
						order.Floor  = floor
						order.Button = types.ButtonType(button)
						
						orderTimedOutChn <- order
					}
				}
			}
		}
	}
}

func CombineHRAInput(
	RxElevInfoChan <-chan LOCAL_ELEVATOR_INFO, 
	RxNewHallRequestChan <-chan BUTTON_INFO, 
	RxFinishedHallOrderChan <-chan BUTTON_INFO,

	TxHRAInputChan chan<- HRAInput){

	currentHRAInput := HRAInput{
		HallRequests: make([NUM_FLOORS][NUM_BUTTONS-1]bool),
}

for i := range currentHRAInput.HallRequests {
    for j := range currentHRAInput.HallRequests[i] {
        currentHRAInput.HallRequests[i][j] = false
    }
}


	for{
		select{
		case newElevInfo <- RxElevInfoChan:
			currentHRAInput.States[newElevInfo.ElevatorID]=newElevInfo			
			TxHRAInputChan <- currentHRAInput

		case newHRequest <-RxNewHallRequestChan:
			if currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button]==0{
				currentHRAInput.HallRequests[newHRequest.Floor][newHRequest.Button]=1
				TxHRAInputChan <- currentHRAInput
			}

		case finishedHOrder <- RxFinishedHallOrderChan:
			currentHRAInput.HallRequests[finishedHOrder.Floor][finishedHOrder.Button]=0
			TxHRAInputChan <- currentHRAInput

		case initInfo <- InitInfoChan:

		default:

		}
	}
}
	

