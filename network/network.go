package network

import (
	"../utilities"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

var client_port int = 30056
var master_port int = 30057
var client_backup_port int = 30058
var master_backup_port int = 30059

const InvalidID utilities.ID = ""

func EncodeMessage(m utilities.Message) []byte {

	result, err := json.Marshal(m)

	if err != nil {
		log.Fatal(err)
	}

	return result
}

func DecodeMessage(b []byte) utilities.Message {
	var result utilities.Message
	err := json.Unmarshal(b, &result)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func getSenderID(sender *net.UDPAddr) utilities.ID {
	return utilities.ID(sender.IP.String())
}

func GetMachineID() utilities.ID {
	ifaces, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	for _, addr := range ifaces {
		if ip_addr, ok := addr.(*net.IPNet); ok && !ip_addr.IP.IsLoopback() {
			if v4 := ip_addr.IP.To4(); v4 != nil {
				return utilities.ID(v4.String())
			}
		}
	}
	return utilities.DEFAULT_IP
}

func listen(socket *net.UDPConn, incoming chan utilities.Packet, quit chan bool) {
	for {
		select {
		case <-quit:
			socket.Close()
			return
		default:
			setDeadLine(socket, time.Now())
			bytes := make([]byte, 2048)
			read_bytes, sender, err := socket.ReadFromUDP(bytes)

			if err == nil {
				incoming <- utilities.Packet{getSenderID(sender), bytes[:read_bytes]}
			}

			if err != nil && !err.(net.Error).Timeout() {
				log.Println(err)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func broadcast(socket *net.UDPConn, to_port int, outgoing chan utilities.Packet, quit chan bool, local bool) {
	address := GetMachineID()
	if !local {
		address = "255.255.255.255"
	}
	bcast_addr := fmt.Sprintf("%s:%d", address, to_port)
	remote, err := net.ResolveUDPAddr("udp", bcast_addr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		packet := <-outgoing
		_, err := socket.WriteToUDP(packet.Data, remote)
		if err != nil {

			if local {

				//Show error
				log.Println(err)

				//Wait
				time.Sleep(1 * time.Second)

				//Reconnect
				broadcast(socket, to_port, outgoing, quit, local)

				//Quit this loop
				break

			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func setDeadLine(socket *net.UDPConn, t time.Time) {

	err := socket.SetReadDeadline(t.Add(time.Millisecond * 2000))

	if err != nil && !err.(net.Error).Timeout() {
		log.Fatal(err)
	}

}

func bind(port int) *net.UDPConn {
	local, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	socket, err := net.ListenUDP("udp", local)
	if err != nil {
		log.Fatal(err)
	}
	return socket
}

func ClientWorker(from_master, to_master chan utilities.Packet, quit chan bool) {
	socket := bind(client_port)
	go listen(socket, from_master, quit)
	broadcast(socket, master_port, to_master, quit, false)
	socket.Close()
}

func MasterWorker(from_client, to_clients chan utilities.Packet, quit chan bool) {
	socket := bind(master_port)
	go listen(socket, from_client, quit)
	broadcast(socket, client_port, to_clients, quit, false)
	socket.Close()
}

func ClientBackupWorker(from_master, to_master chan utilities.Packet, quit chan bool) {
	socket := bind(client_backup_port)
	listen(socket, from_master, quit)
	socket.Close()
}

func MasterBackupWorker(from_client, to_clients chan utilities.Packet, quit chan bool) {
	socket := bind(master_backup_port)
	broadcast(socket, client_backup_port, to_clients, quit, true)
	socket.Close()
}
