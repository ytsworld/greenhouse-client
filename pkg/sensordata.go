package greenhouse

// Data contains sensor data in case of success or a error message that cen be used on server side for alerting
type Data struct {
	Success                bool
	Message                string
	Temperature            float32
	Humidity               float32
	SoilMoistureResistance int
	UnixTimestampUTC       int64
}
