package main

import (
	"time"
)

type persistMessage struct {
	Timestamp time.Time
	Rssi      string
	Mac       string
	Name      string
}
