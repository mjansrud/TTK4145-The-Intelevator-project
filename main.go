package main

/*

Explanation of packages:
-----------------------

Main: The executing package upon program start.
Startup: Guarantees a successfull startup, creates all necessary threads and a backup terminal
Driver: Is a set of functions to control the elevator
Hardware: Lowest level interface to interact with the elvator
Events: Event loops, the brain of the system
Orders: Handles everything regarding the orders.
Network: The network module allowing communication between elevators.
States: Tracks the local elevator states and manages the connected elevators.
Utilities: Adds the structs and definitions used throughout the program.

*/

import (
	"./startup"
	"./states"
	"flag"
	"time"
)

var filename string = "Main -"

func main() {

	//Get default master variable - boolean false
	flag_master := states.IsMaster()

	//Startup flags - aka "go run -master"
	flag.BoolVar(&flag_master, "master", false, "Start as master")
	flag.Parse()

	//Update state machine
	states.SetMaster(flag_master)

	//If process is master on local machine
	if states.IsMaster() {

		startup.InitMaster()

	}

	//If process is backup
	if !states.IsMaster() {

		startup.InitBackup()

	}

	//Prevent the system from stopping
	for {

		time.Sleep(500 * time.Millisecond)

	}

}
