package startup

import (
	"../driver"
	"../events"
	"../network"
	"../orders"
	"../states"
	"../utilities"
	"fmt"
	"os/exec"
)

var filename string = "Startup -"

//Backup channels for localhost connection
var channel_backup_write = make(chan utilities.Packet)
var channel_backup_listen = make(chan utilities.Packet)

//Toggle channels
var channel_backup_init_master = make(chan bool)
var channel_backup_quit = make(chan bool)

//Main channels for global connection
var channel_listen = make(chan utilities.Packet)
var channel_write = make(chan utilities.Packet)

//Channels for orders
var channel_floor_poll = make(chan utilities.Floor)
var channel_order_poll = make(chan utilities.Order)

//Functions
func CreateBackup() {

	//Run shell
	out, err := exec.Command("sh", "-c", "gnome-terminal -e 'go run main.go'").Output()

	if err != nil {
		fmt.Println(filename, "Error - ", out, err)
	}
}

//Initiate
func InitBackup() {

	//Log
	fmt.Println(filename, "Starting the intelevator as backup!")

	//Network
	go network.ClientBackupWorker(channel_backup_listen, channel_backup_write, channel_backup_quit)

	//Heartbeat
	go events.HeartbeatChecker(channel_backup_quit, channel_backup_init_master)
	go events.HeartbeatListener(channel_backup_init_master, channel_backup_listen)

	//Initiate a new master if neccessary
	go RestoreMaster()
}

func InitMaster() {

	//Log
	fmt.Println(filename, "Starting the intelevator as master!")
	fmt.Println(filename, "Machine id :", network.GetMachineID())

	//Driver
	driver.Init()

	//Network communication
	go network.MasterBackupWorker(channel_backup_listen, channel_backup_write, channel_backup_quit)
	go network.MasterWorker(channel_listen, channel_write, channel_backup_quit)
	go network.ClientWorker(channel_listen, channel_write, channel_backup_quit)

	//Backup
	CreateBackup()

	//Update state machine
	states.SetMaster(true)

	//Events
	go events.PollFloor(channel_floor_poll)
	go events.PollOrder(channel_order_poll)
	go orders.Handle(channel_floor_poll, channel_order_poll, channel_write)

	//Listener and broadcaster
	go events.NetworkLoop(channel_write)
	go events.NetworkListener(channel_listen, channel_write)
	go events.HeartbeatLoop(channel_backup_write)
	go events.StatusChecker(channel_write)

	//Catch ctrl+c termination - stop elevator
	go events.ExitHandler()
}

func RestoreMaster() {

	//Run loop forever
	for {

		//Check if we are master
		if !states.IsMaster() {

			select {
			case <-channel_backup_init_master:

				//We are no longer a backup
				InitMaster()

				break
			}
		}
	}

}
