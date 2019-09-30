package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/d2r2/go-dht"
	sal "github.com/salrashid123/oauth2/google"
	greenhouse "github.com/ytsworld/greenhouse-client/pkg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/spi"
	"gobot.io/x/gobot/platforms/raspi"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	ledPin                 = "8"
	soilMoistureSPIChannel = 0
	greenhouseReceiverURL  = "https://europe-west1-yt-dev-242017.cloudfunctions.net/greenhouse-receiver"
	greenhouseReceiverAPI  = "/api/v1/greenhouse"
)

var (
	timeout = time.Duration(30 * time.Second)
	ctx     = context.Background()
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
			data.Success = true

			led.Off()

			err = sendData(&data)
			if err != nil {
				data.Message = fmt.Sprintf("Error sending data to server. %s", err)
				reportError(&data, led)
			}

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

func sendData(data *greenhouse.Data) error {

	client, err := getAuthorizedClient()
	if err != nil {
		return err
	}

	jsonPayload, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	fmt.Printf("Sensor data json: %s\n", jsonPayload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", greenhouseReceiverURL, greenhouseReceiverAPI), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("Unexpected response from server: %d - %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func reportSuccess(data *greenhouse.Data, led *gpio.LedDriver) {
	indicateSuccessOnLed(led)
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

// getAuthorizedClient returns a http client that identifies against cloud function with an identity token
func getAuthorizedClient() (*http.Client, error) {
	scopes := "https://www.googleapis.com/auth/userinfo.email"
	creds, err := google.FindDefaultCredentials(ctx, scopes)
	if err != nil {
		return nil, err
	}
	targetAudience := greenhouseReceiverURL

	idTokenSource, err := sal.IdTokenSource(
		sal.IdTokenConfig{
			Credentials: creds,
			Audiences:   []string{targetAudience},
		},
	)
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: idTokenSource,
		},
	}

	return client, nil

}
