package sensor

import "time"

type TemperatureSensorStatus struct {
	is_available       bool
	temperature        float64
	last_status_change time.Time
}

func (t *TemperatureSensorStatus) New() TemperatureSensorStatus {
	return TemperatureSensorStatus{is_available: false, temperature: 0, last_status_change: time.Now()}
}

func (t *TemperatureSensorStatus) Update(status bool, temp float64) {
	t.is_available = status
	t.temperature = temp
	t.last_status_change = time.Now()
}

func (t *TemperatureSensorStatus) IsAvailable() bool {
	return t.is_available
}

func (t *TemperatureSensorStatus) Temperature() float64 {
	return t.temperature
}

func (t *TemperatureSensorStatus) LastStatusChange() time.Time {
	return t.last_status_change
}

type TemperatureSensorReading struct {
	Status      bool
	Temperature float64
}
