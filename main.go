package main

import (
	network "project-group-74\network"
    dec "project-group-74\decision"
    elev_control "project-group-74\elev_control"
)

func main() {

    elev_init()
        decide_master()
    ElevControlChns := make_elev_control_chns()
    NetworkChns := make_network_chns()
    go elev_control.Elev_run(ElevControlChns)
    go network.Network_run(NetworkChns)
    go elevio.PollButtons(ElevControlChns.NewBtnpress)
    go elevio.PollNewFloor(ElevControlChns.NewFloor)


}