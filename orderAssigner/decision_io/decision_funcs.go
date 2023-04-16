package decision_io

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"project-group-74/localTypes"
)

func NewAllFalseHRAInput() localTypes.HRAInput {
	output := localTypes.HRAInput{}
	for i := range output.HallRequests {
		for j := range output.HallRequests[i] {
			output.HallRequests[i][j] = false
		}
	}
	output.States = make(map[string]localTypes.HRAElevState)
	return output
}

func ReassignOrders(newHRAInput localTypes.HRAInput, hraExecutable string) map[string]localTypes.HMATRIX {
	jsonBytes, err := json.Marshal(newHRAInput)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
	}

	ret, err := exec.Command("orderAssigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()

	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
	}

	output := map[string]localTypes.HMATRIX{}
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
	}
	return output
}

func LocalState2HRASTATE(newElevInfo localTypes.LOCAL_ELEVATOR_INFO) localTypes.HRAElevState {
	output := localTypes.HRAElevState{

		State:       elevStateStrings[newElevInfo.State],
		Floor:       newElevInfo.Floor,
		Direction:   motorDirStrings[newElevInfo.Direction],
		CabRequests: newElevInfo.CabCalls,
	}
	return output
}
