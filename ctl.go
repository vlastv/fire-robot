package main

import (
	"fmt"
	"log"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/api"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"

	"flag"

	local "./drivers/gpio"
)

func main() {
	debug := flag.Bool("debug", false, "Debug API")
	doorOpenTemp := flag.Int("temp-open-door", 30, "Temperature for trigger open door")
	flapOpenTemp := flag.Int("temp-open-flap", 40, "Temperature for trigger open flap")
	allCloseTemp := flag.Int("temp-close-all", 24, "Temperature for trigget class all")
	dhtInterval := flag.String("interval-poll-sensor", "5m", "Interval polling DHT")

	flag.Parse()

	fmt.Println(
		time.Microsecond,
		time.Millisecond,
		time.Nanosecond,
	)

	master := gobot.NewMaster()

	a := api.NewAPI(master)

	if *debug {
		a.Debug()
	}

	a.Start()

	r := raspi.NewAdaptor()

	m1 := gpio.NewMotorDriver(r, "")
	m1.SetName("Door Motor")
	m1.ForwardPin = "7"   // 4
	m1.BackwardPin = "11" // 17

	m2 := gpio.NewMotorDriver(r, "")
	m2.SetName("Flap Motor")
	m2.ForwardPin = "13"  // 27
	m2.BackwardPin = "15" // 22

	s := local.NewDHTSensorDriver(r, "16") //23
	s.SetName("Sensor")
	if d, err := time.ParseDuration(*dhtInterval); err != nil {
		s.PollInterval(d)
	}

	f := NewFire(m1, m2, s)

	work := func() {
		s.On(gpio.Data, func(data interface{}) {
			t := data.(float32)
			log.Printf("Current temperature %.1f", t)

			if t >= float32(*flapOpenTemp) {
				f.OpenFlap()
			}

			if t >= float32(*doorOpenTemp) {
				f.OpenDoor()
			}

			if t > 0.0 && t <= float32(*allCloseTemp) {
				f.CloseDoor()
				f.CloseFlap()
			}
		})

		s.On(gpio.Error, func(data interface{}) {
			log.Print("Error DHT")
			log.Println(data)
		})
	}

	master.AddRobot(gobot.NewRobot("fire",
		[]gobot.Connection{r},
		[]gobot.Device{m1, m2, s},
		work,
	))

	master.Start()
}
