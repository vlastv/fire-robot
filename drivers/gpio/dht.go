package gpio

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"

	"../../dht"
)

type digitalReadWriter interface {
	gobot.Adaptor
	DigitalRead(string) (val int, err error)
	DigitalWrite(string, byte) (err error)
}

// DHTSensorDriver represents a Digital-output relative humidity & temperature sensor/module
type DHTSensorDriver struct {
	connection digitalReadWriter
	name       string
	pin        string
	interval   time.Duration
	halt       chan bool
	retries    int
	gobot.Eventer
	gobot.Commander
}

// NewDHTSensorDriver returns a new DHTSensorDriver given a DigitalWrite/DigitalRead and pin
func NewDHTSensorDriver(a digitalReadWriter, pin string) *DHTSensorDriver {
	d := &DHTSensorDriver{
		connection: a,
		name:       gobot.DefaultName("DHTSensor"),
		pin:        pin,
		interval:   1 * time.Minute,
		halt:       make(chan bool),
		retries:    15,
		Eventer:    gobot.NewEventer(),
		Commander:  gobot.NewCommander(),
	}

	d.AddEvent(gpio.Error)
	d.AddEvent(gpio.Data)
	d.AddCommand("Read", func(params map[string]interface{}) interface{} {
		v, err := d.RetryRead()

		return map[string]interface{}{"val": v, "err": err}
	})

	return d
}

// Name returns the DHTSensorDriver name
func (d *DHTSensorDriver) Name() string { return d.name }

// SetName sets the DHTSensorDriver name
func (d *DHTSensorDriver) SetName(n string) { d.name = n }

// Start implements the Driver interface
func (d *DHTSensorDriver) Start() error {
	var value float32

	go func() {
		for {
			newValue, err := d.RetryRead()
			if err != nil {
				d.Publish(d.Event(gpio.Error), err)
			} else if newValue != value && newValue != -1 {
				value = newValue
				d.Publish(d.Event(gpio.Data), value)
			}

			select {
			case <-time.After(d.interval):
			case <-d.halt:

				return
			}
		}
	}()

	return nil
}

// Pin returns the DHTSensorDriver pin
func (d *DHTSensorDriver) Pin() string { return d.pin }

// Connection returns the DHTSensorDriver Connection
func (d *DHTSensorDriver) Connection() gobot.Connection { return d.connection }

// Halt implements the Driver interface
func (d *DHTSensorDriver) Halt() error {
	d.halt <- true

	return nil
}

func (d *DHTSensorDriver) Read() (float32, error) {
	w, err := dht.Read(d.connection, d.pin)
	if err != nil {
		return 0.0, err
	}

	return w.Temperature, nil
}

// RetryRead repeats reading until the first successful
func (d *DHTSensorDriver) RetryRead() (float32, error) {
	for i := 0; i < d.retries; i++ {
		if v, err := d.Read(); err == nil {
			return v, err
		}
		time.Sleep(3 * time.Second)
	}

	return 0.0, dht.TimeoutError
}

// PollInterval sets interval of polling DHTSensor
func (d *DHTSensorDriver) PollInterval(i time.Duration) {
	d.interval = i
}
