package dht

import (
	"errors"
	"time"
)

const (
	COLLECTING_PERIOD  = 2 * time.Second
	LOGICAL_1_TRESHOLD = 50 * time.Microsecond
)

var (
	ChecksumError    = errors.New("checksum error")
	HumidityError    = errors.New("humidity range error")
	TemperatureError = errors.New("temperature range error")
	TimeoutError     = errors.New("timeout error")
)

const (
	HIGH = 1
	LOW  = 0
)

type DigitalReadWriter interface {
	DigitalRead(string) (val int, err error)
	DigitalWrite(string, byte) (err error)
}

func Read(c DigitalReadWriter, p string) (Weather, error) {
	weather := Weather{}

	pulse, err := dhtDigitalRead(c, p)
	if err != nil {
		return weather, err
	}

	bytes, err := dhtParse(pulse)
	if err != nil {
		return weather, err
	}

	var (
		humidity    uint16
		temperature uint16
	)

	humidity = uint16(bytes[0])<<8 + uint16(bytes[1])
	hf := float32(humidity) / 10

	if hf < 0 || hf > 100 {
		return weather, HumidityError
	}

	weather.Humidity = hf

	temperature = uint16(bytes[2])<<8 + uint16(bytes[3])
	if temperature&0x8000 > 0 {
		temperature ^= 0x8000
	}
	tf := float32(temperature) / 10

	if tf < -40.0 || tf > 80 {
		return weather, TemperatureError
	}

	weather.Temperature = tf

	return weather, nil
}

type Weather struct {
	Temperature float32
	Humidity    float32
}

func dhtDigitalRead(p DigitalReadWriter, pin string) ([]bool, error) {
	s := make([]bool, 40)
	threshold := 60 * time.Microsecond
	limit := 32000

	readHigh := func() (time.Duration, error) {
		for i := 0; i < limit; i++ {
			v, err := p.DigitalRead(pin)
			if err != nil {
				return 1, err
			}

			if v == 1 {
				break
			}
		}

		t := time.Now()
		for i := 0; i < limit; i++ {
			v, err := p.DigitalRead(pin)
			if err != nil {
				return 2, err
			}

			if v == 0 {
				return time.Since(t), nil
			}
		}

		return 3, TimeoutError
	}

	if err := p.DigitalWrite(pin, LOW); err != nil {
		return s, err
	}
	time.Sleep(1 * time.Millisecond)

	if err := p.DigitalWrite(pin, HIGH); err != nil {
		return s, err
	}

	if _, err := readHigh(); err != nil {
		return s, err
	}

	for i := 0; i < 40; i++ {
		v, err := readHigh()
		if err != nil {
			return s, err
		}

		s[i] = v > threshold
	}

	return s, nil
}

func dhtParse(pulse []bool) ([]uint8, error) {
	bytes := make([]uint8, 5)

	for i := range bytes {
		for j := 0; j < 8; j++ {
			bytes[i] <<= 1
			if pulse[i*8+j] {
				bytes[i] |= 0x01
			}
		}
	}

	if err := dhtChecksum(bytes); err != nil {
		return bytes, nil
	}

	return bytes, nil
}

func dhtChecksum(bytes []uint8) error {
	var sum uint8

	for i := 0; i < 4; i++ {
		sum += bytes[i]
	}

	if sum != bytes[4] {
		return ChecksumError
	}

	return nil
}
