package greenhouse

// Data contains collected sensor data in case of success or a error message that can be used on server side for alerting
type Data struct {
	Success                bool    `json:"success"`
	UnixTimestampUTC       int64   `json:"unixTimestampUTC"`
	Message                string  `json:"message,omitempty"`
	Temperature            float32 `json:"temperature,omitempty"`
	Humidity               float32 `json:"humidity,omitempty"`
	SoilMoistureResistance int     `json:"soilMoistureResistance,omitempty"`
}
