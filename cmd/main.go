package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	soilMoistureSPIChannel = 0
	greenhouseReceiverAPI  = "/api/v1/greenhouse"
)

var (
	timeout               = time.Duration(30 * time.Second)
	ctx                   = context.Background()
	r                     *raspi.Adaptor
	adc                   *spi.MCP3008Driver
	progressPin           string // physical pin - LED pin for indication of measurement progress and success
	errorPin              string // physical pin - LED pin to indicate errors
	dht22Pin              int    // GPIO pin - digital output of DHT22
	progressLed           *gpio.LedDriver
	errorLed              *gpio.LedDriver
	greenhouseReceiverURL string
	measureEveryDuration  int // number of seconds to between measurements
)

func init() {
	greenhouseReceiverURL = os.Getenv("GREENHOUSE_RECEIVER_URL")
	if greenhouseReceiverURL == "" {
		panic(fmt.Errorf("Expecting server url in env var GREENHOUSE_RECEIVER_URL"))
	}
	progressPin = os.Getenv("PROGRESS_LED_PHYSICAL_PIN")
	errorPin = os.Getenv("ERROR_LED_PHYSICAL_PIN")
	if errorPin == "" || progressPin == "" {
		panic(fmt.Errorf("Missing env vars PROGRESS_LED_PIN or ERROR_LED_PIN containing physical pin with leds"))
	}

	var err error
	dht22Pin, err = strconv.Atoi(os.Getenv("DHT22_DATA_GPIO_PIN"))
	if err != nil {
		panic(fmt.Errorf("env var DHT22_DATA_GPIO_PIN does not valid GPIO pin"))
	}
	if dht22Pin == 0 {
		panic(fmt.Errorf("Missing env var DHT22_DATA_GPIO_PIN containing the GPIO pin with DHT22 digital output. %s", err))
	}

	measureEveryDuration, err = strconv.Atoi(os.Getenv("MEASURE_EVERY_SECONDS"))
	if err != nil {
		panic(fmt.Errorf("env var MEASURE_EVERY_SECONDS does not contain a valid amount of seconds. %s", err))
	}
	if measureEveryDuration == 0 {
		panic(fmt.Errorf("Missing env var MEASURE_EVERY_SECONDS does not contain a valid amount of seconds"))
	}

	r = raspi.NewAdaptor()
	adc = spi.NewMCP3008Driver(r)
	adc.Start()
	progressLed = gpio.NewLedDriver(r, progressPin)
	errorLed = gpio.NewLedDriver(r, errorPin)
	progressLed.Off()
	errorLed.Off()
}

func main() {

	// Use led to show that the program has started and check that the leds are connected properly
	onStartSequence()

	work := func() {
		gobot.Every(time.Duration(measureEveryDuration)*time.Second, func() {
			data := greenhouse.Data{}
			now := time.Now()
			data.UnixTimestampUTC = now.Unix()
			progressLed.On()

			resistance, err := measureSoilMoisture(adc)
			if err != nil {
				data.Success = false
				data.Message = fmt.Sprintf("%s", err)
				trySendErrorReport(&data)
				return

			}
			data.SoilMoistureResistance = resistance

			temperature, humidity, err := measureTemp()

			if err != nil {
				data.Success = false
				data.Message = fmt.Sprintf("%s", err)
				trySendErrorReport(&data)
				return
			}
			data.Temperature = temperature
			data.Humidity = humidity
			data.Success = true

			progressLed.Off()

			err = sendData(&data)
			if err != nil {
				fmt.Printf("Error sending data to server. %s\n", err)
				indicateError()
				return
			}

			indicateSuccess()

		})
	}

	robot := gobot.NewRobot("greenhouse",
		[]gobot.Connection{r},
		[]gobot.Device{progressLed, errorLed, adc},
		work,
	)

	robot.Start()
}

func onStartSequence() {
	progressLed.On()
	errorLed.On()
	time.Sleep(2 * time.Second)
	progressLed.Off()
	errorLed.Off()
	time.Sleep(500 * time.Millisecond)
	progressLed.On()
	time.Sleep(1 * time.Second)
	progressLed.Off()

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

	fmt.Printf("Successfully sent data to server.\n")

	return nil
}

func trySendErrorReport(data *greenhouse.Data) {
	fmt.Printf("ERROR: %s\n", data.Message)
	fmt.Printf("Trying to send error report to server...\n")
	err := sendData(data)
	if err != nil {
		fmt.Printf("ERROR sending error report to server: %v\n", err)
	} else {
		fmt.Printf("Error report was sent\n")
	}
	indicateError()
}

func indicateSuccess() {
	errorLed.Off()
	progressLed.Off()
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 2; i++ {
		progressLed.On()
		time.Sleep(300 * time.Millisecond)
		progressLed.Off()
		time.Sleep(200 * time.Millisecond)
	}
}

func indicateError() {
	errorLed.Off()
	progressLed.Off()
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 5; i++ {
		errorLed.On()
		time.Sleep(100 * time.Millisecond)
		errorLed.Off()
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

// measureTemp requests the DHT22 sensor for temperature and humidity data
func measureTemp() (temperature float32, humidity float32, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Error while requesting data from DHT22. %v", e)
		}
	}()

	temperature, humidity, _, err = dht.ReadDHTxxWithRetry(dht.DHT22, dht22Pin, false, 5)

	return
}

// getAuthorizedClient returns a http client that identifies against cloud function with an identity token
// Path to service account credential json file is taken from environment variable GOOGLE_APPLICATION_CREDENTIALS
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
