package synchroniser

import (
	"../global"
	"../network"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const filename = "backupFile"

// Loads backup from a file (if it exists). Resend orders to network if backup not empty
func runBackup(outgoingMsg chan<- global.Message) {
	var backup queue
	backup.loadFromDisk(filename)

	// Resend all orders found on loaded backup file:
	if !backup.isEmpty() {
		for floor := 0; floor < global.NumFloors; floor++ {
			for button := 0; button < global.NumButtons; button++ {
				if backup.isOrder(floor, button) {
					if button == global.BtnInside {
						AddLocalOrder(floor, button)
					} else {
						if network.UDP_Alive() {
							outgoingMsg <- global.Message{Category: global.Msg_NewOrder, Floor: floor, Button: button}
						}
					}
				}
			}
		}
	}
	go func() {
		for {
			<-takeBackup
			if err := local.saveToDisk(filename); err != nil {
				log.Println(global.ColorR, err, global.ColorN)
			}
		}
	}()
}

// Saves a queue to the disk with given filename, returns error message if error
func (q *queue) saveToDisk(filename string) error {
	data, err := json.Marshal(&q)
	if err != nil {
		log.Println(global.ColorR, "json.Marshal() error: Failed to backup.", global.ColorN)
		return err
	}
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		log.Println(global.ColorR, "ioutil.WriteFile() error: Failed to backup.", global.ColorN)
		return err
	}
	return nil
}

// saves content from disk from "filename" if the file exists. 
func (q *queue) loadFromDisk(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		log.Println(global.ColorG, "Backup file found, processing...", global.ColorN)

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println(global.ColorR, "loadFromDisk() error: Failed to read file.", global.ColorN)
		}
		if err := json.Unmarshal(data, q); err != nil {
			log.Println(global.ColorR, "loadFromDisk() error: Failed to Unmarshal.", global.ColorN)
		}
	}
	return nil
}