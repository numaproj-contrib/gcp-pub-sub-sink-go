package mocks

import (
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	Data string
}

func (p Payload) Keys() []string {
	return nil
}

func (p Payload) Value() []byte {
	return []byte(p.Data)
}
func (p Payload) EventTime() time.Time {
	return time.Now()
}
func (p Payload) Watermark() time.Time {
	return time.Now()
}
func (p Payload) ID() string {
	return uuid.New().String()
}
