#! /bin/sh

GOOGLE_APPLICATION_CREDENTIALS="/home/pi/greenhouse-client.sa.json"
export GOOGLE_APPLICATION_CREDENTIALS
GREENHOUSE_RECEIVER_URL="{{ greenhouse_recv_url }}"
export GREENHOUSE_RECEIVER_URL
ERROR_LED_PHYSICAL_PIN={{ greenhouse_client_error_led_physical_pin }}
export ERROR_LED_PHYSICAL_PIN
PROGRESS_LED_PHYSICAL_PIN={{ greenhouse_client_progress_led_physical_pin }}
export PROGRESS_LED_PHYSICAL_PIN
DHT22_DATA_GPIO_PIN={{ greenhouse_client_dht22_gpio_pin }}
export DHT22_DATA_GPIO_PIN
MEASURE_EVERY_SECONDS={{ greenhouse_client_measure_every_seconds }}
export MEASURE_EVERY_SECONDS

/home/pi/greenhouse-client
