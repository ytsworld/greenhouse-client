package main

import (
	"fmt"
	"time"

	greenhouse "github.com/YT84/greenhouse-client/pkg"
	"github.com/d2r2/go-dht"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/spi"
	"gobot.io/x/gobot/platforms/raspi"
)

const (
	ledPin                 = "8"
	soilMoistureSPIChannel = 0
)

func main() {
	r := raspi.NewAdaptor()
	adc := spi.NewMCP3008Driver(r)
	adc.Start()
	led := gpio.NewLedDriver(r, ledPin)
	led.Off()

	work := func() {
		gobot.Every(20*time.Second, func() {
			data := greenhouse.Data{}
			now := time.Now()
			data.UnixTimestampUTC = now.Unix()
			led.On()

			resistance, err := measureSoilMoisture(adc)
			if err != nil {
				data.Success = false
				data.Message = fmt.Sprintf("%s", err)
				reportError(&data, led)
				return

			}
			data.SoilMoistureResistance = resistance

			temperature, humidity, err := measureTemp()

			if err != nil {
				data.Success = false
				data.Message = fmt.Sprintf("%s", err)
				reportError(&data, led)
				return
			}
			data.Temperature = temperature
			data.Humidity = humidity

			led.Off()

			data.Message = fmt.Sprintf("Temperature = %v*C, Humidity = %v%%, Soil Moisture: %d\n",
				temperature, humidity, resistance)

			reportSuccess(&data, led)

		})
	}

	robot := gobot.NewRobot("greenhouse",
		[]gobot.Connection{r},
		[]gobot.Device{led, adc},
		work,
	)

	robot.Start()
}

func reportSuccess(data *greenhouse.Data, led *gpio.LedDriver) {
	indicateSuccessOnLed(led)
	fmt.Printf("Success: %s\n", data.Message)
}

func reportError(data *greenhouse.Data, led *gpio.LedDriver) {
	indicateErrorOnLed(led)
	fmt.Printf("ERROR: %s\n", data.Message)
}

func indicateSuccessOnLed(led *gpio.LedDriver) {
	led.Off()
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 2; i++ {
		led.On()
		time.Sleep(300 * time.Millisecond)
		led.Off()
		time.Sleep(200 * time.Millisecond)
	}
}

func indicateErrorOnLed(led *gpio.LedDriver) {
	led.Off()
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 5; i++ {
		led.On()
		time.Sleep(100 * time.Millisecond)
		led.Off()
		time.Sleep(50 * time.Millisecond)
	}
}

// MCP3008Driver converts the analog signal to a digit between 1 and 1024
// 1 means probably the pins are connected by cable, 1024 means no connection
// Lets do interpretation on server side
func measureSoilMoisture(adc *spi.MCP3008Driver) (resistance int, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Error while requesting data from MCP3008. %v", e)
		}
	}()

	resistance, err = adc.Read(soilMoistureSPIChannel)

	return resistance, err
}

func measureTemp() (temperature float32, humidity float32, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Error while requesting data from DHT22. %v", e)
		}
	}()

	temperature, humidity, _, err = dht.ReadDHTxxWithRetry(dht.DHT22, 17, false, 5)

	return
}
