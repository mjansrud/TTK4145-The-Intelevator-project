package states

//Imports
import (
	"../utilities"
	"fmt"
	"sync"
)

var filename string = "State machine -"
var connected_ip utilities.ID = "0.0.0.0"
var mutex sync.Mutex
var master bool = false
var state int = utilities.STATE_STARTUP
var direction int = utilities.INVALID
var floor int = 0
var connected int = utilities.DISCONNECTED
var priority int = 0

//Functions
func IsMaster() bool {

	mutex.Lock()
	defer mutex.Unlock()

	return master
}

func IsConnected() bool {

	mutex.Lock()
	defer mutex.Unlock()

	if connected == utilities.CONNECTED {
		return true
	}

	return false
}

func GetState() int {

	mutex.Lock()
	defer mutex.Unlock()

	return state
}

func GetFloor() int {

	mutex.Lock()
	defer mutex.Unlock()

	return floor
}

func GetDirection() int {

	mutex.Lock()
	defer mutex.Unlock()

	return direction
}

func GetPriority() int {

	mutex.Lock()
	defer mutex.Unlock()

	return priority
}

func GetConnectedIp() utilities.ID {

	mutex.Lock()
	defer mutex.Unlock()

	return connected_ip
}

//Setters
func SetMaster(local_master bool) {

	mutex.Lock()
	master = local_master
	mutex.Unlock()

}

//Setters
func SetConnection(local_connected int) {

	mutex.Lock()
	if connected != local_connected {

		//Set connection state
		connected = local_connected

		fmt.Println(filename, "Setting new connectivity: ", connected)

	}
	mutex.Unlock()

}

func SetState(local_state int) {

	mutex.Lock()
	if state != local_state {

		//Set state
		state = local_state

		fmt.Println(filename, "Setting new state: ", PrintState())

	}
	mutex.Unlock()

}

func PrintState() string {

	switch state {
	case utilities.STATE_STARTUP:
		return "startup"
	case utilities.STATE_IDLE:
		return "idle"
	case utilities.STATE_RUNNING:
		return "running"
	case utilities.STATE_EMERGENCY:
		return "emergency"
	case utilities.STATE_DOOR_OPEN:
		return "door open"
	case utilities.STATE_DOOR_CLOSED:
		return "door closed"
	}

	return "invalid"
}

func SetFloor(local_floor int) {

	mutex.Lock()
	if floor != local_floor {

		floor = local_floor
		fmt.Println(filename, "Setting new floor: ", floor)

	}
	mutex.Unlock()

}

func SetDirection(local_direction int) {

	mutex.Lock()
	if direction != local_direction {

		direction = local_direction
		fmt.Println(filename, "Setting new direction: ", direction)

	}
	mutex.Unlock()

}

func SetPriority(local_priority int) {

	mutex.Lock()
	if priority != local_priority {

		priority = local_priority
		fmt.Println(filename, "Setting new priority: ", priority)

	}
	mutex.Unlock()

}

func SetConnectedIp(local_ip utilities.ID) {

	mutex.Lock()
	if connected_ip != local_ip {

		connected_ip = local_ip

	}
	mutex.Unlock()

}
