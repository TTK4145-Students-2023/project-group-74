import (
	"Driver-go/elevio"
	"fmt"
)

design desicions:
--Change motordirn before going to state Moving or in state Moving

state Idle
If queue is empty -> state Idle
if queue not empty -> state Moving

state Moving	
	first_in_queue-> target floor
		if targetfloor =/</> currentfloor
			SetMotorDirection=0/1
	for MotorDirection=!2
		





// for f := 0; f < numFloors; f++ {
// 	for b := elevio.ButtonType(0); b < 3; b++ {
// 		if OrderMatrix[f][b]==1{
// 		elevio.SetButtonLamp(b, f, true)
// 		}
// 	}
// }


var OrderMatrix [3][4] int


select {
case a := <-drv_buttons:
	order_matrix[a.Button][a.Floor]=1;
	fmt.Printf("%+v\n", a)
	elevio.SetButtonLamp(a.Button, a.Floor, true)

case a := <-drv_floors:
	fmt.Printf("%+v\n", a)
	if a == numFloors-1 {
		d = elevio.MD_Down
	} else if a == 0 {
		d = elevio.MD_Up
	}
	elevio.SetMotorDirection(d)

}