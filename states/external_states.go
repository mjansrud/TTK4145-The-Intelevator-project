package states

//Imports
import (
	"../utilities"
	"fmt"
)

var elevators []utilities.Status

func InsertExternalElevator(elevator utilities.Status) {

	fmt.Println(filename, "Inserting elevator", elevator.Elevator)

	mutex.Lock()
	elevators = append(elevators, elevator)
	mutex.Unlock()

}

func RemoveExternalElevator(elevator utilities.Status) {

	local_elevators := GetExternalElevators()

	//Check if we have elevators
	if len(local_elevators) > 0 {

		for index := range local_elevators {
			if local_elevators[index].Elevator == elevator.Elevator {

				fmt.Println(filename, "Removing elevator")

				mutex.Lock()
				elevators = append(elevators[:index], elevators[index+1:]...)
				mutex.Unlock()

				//Prevent out of bounds and panic
				break
			}
		}
	}

}

func UpdateExternalElevator(new utilities.Status) {

	found := false
	local_elevators := GetExternalElevators()

	//Check if we have elevators
	if len(local_elevators) > 0 {

		for index := range local_elevators {

			//Check if elevator exists
			if local_elevators[index].Elevator == new.Elevator {

				mutex.Lock()
				//Update elevator
				elevators[index].State = new.State
				elevators[index].Floor = new.Floor
				elevators[index].Direction = new.Direction
				elevators[index].Time = new.Time
				mutex.Unlock()

				found = true
			}
		}
	}

	if !found {

		InsertExternalElevator(new)

	}
}

//
func GetExternalElevators() []utilities.Status {

	mutex.Lock()
	defer mutex.Unlock()

	//Create a copy - preventing data race
	e := make([]utilities.Status, len(elevators), len(elevators))

	//Need to manually copy all variables - Library "copy" function will not work
	for id, elem := range elevators {
		e[id] = elem
	}

	return e

}

func CheckExternalElevatorExists(elevator utilities.Status) bool {

	local_elevators := GetExternalElevators()

	//Check if we have elevators
	if len(local_elevators) > 0 {
		for index := range local_elevators {
			if local_elevators[index].Elevator == elevator.Elevator {
				return true
			}
		}
	}

	return false

}
