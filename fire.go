package main

import (
	"log"
	"time"

	local "./drivers/gpio"
	"gobot.io/x/gobot/drivers/gpio"
)

const (
	open  = "open"
	close = "close"
)

type Fire struct {
	door      *gpio.MotorDriver
	flap      *gpio.MotorDriver
	sensor    *local.DHTSensorDriver
	doorState string
	flapState string
	doorTimer *time.Timer
	flapTimer *time.Timer
}

func NewFire(door, flap *gpio.MotorDriver, sensor *local.DHTSensorDriver) *Fire {
	f := &Fire{
		door:   door,
		flap:   flap,
		sensor: sensor,
	}

	return f
}

func (f *Fire) OpenDoor() {
	ok, err := f.ToggleDoor(open)
	if err != nil {
		log.Println("OpenDoor", err)
		return
	}

	if ok {
		log.Println("Door is open")
	} else {
		log.Println("Door is already open")
	}
}

func (f *Fire) CloseDoor() {
	ok, err := f.ToggleDoor(close)
	if err != nil {
		log.Println("CloseDoor", err)
		return
	}

	if ok {
		log.Println("Door is closed")
	} else {
		log.Println("Door is already closed")
	}
}

func (f *Fire) ToggleDoor(state string) (bool, error) {
	if f.doorState == state {
		return false, nil
	}

	if f.doorTimer != nil {
		f.doorTimer.Stop()
	}

	if state == open {
		if err := f.door.Direction("forward"); err != nil {
			log.Println(f.door)
			return false, err
		}
	} else if state == close {
		if err := f.door.Direction("backward"); err != nil {
			log.Println(f.door)
			return false, err
		}
	}

	f.doorState = state

	f.doorTimer = time.AfterFunc(time.Minute, func() {
		f.door.Direction("none")
	})

	return true, nil
}

func (f *Fire) OpenFlap() {
	ok, err := f.ToggleFlap(open)
	if err != nil {
		log.Println("OpenFlap", err)
		return
	}

	if ok {
		log.Println("Flap is open")
	} else {
		log.Println("Flap is already open")
	}
}

func (f *Fire) CloseFlap() {
	ok, err := f.ToggleFlap(close)
	if err != nil {
		log.Println("CloseFlap", err)
		return
	}

	if ok {
		log.Println("Flap is closed")
	} else {
		log.Println("Flap is already closed")
	}
}

func (f *Fire) ToggleFlap(state string) (bool, error) {
	if f.flapState == state {
		return false, nil
	}

	if f.flapTimer != nil {
		f.flapTimer.Stop()
	}

	if state == open {
		if err := f.flap.Direction("forward"); err != nil {
			log.Println(f.flap)
			return false, err
		}
	} else if state == close {
		if err := f.flap.Direction("backward"); err != nil {
			log.Println(f.flap)
			return false, err
		}
	}

	f.flapState = state

	f.flapTimer = time.AfterFunc(time.Minute, func() {
		f.flap.Direction("none")
	})

	return true, nil
}
