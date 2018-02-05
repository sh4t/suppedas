package main

import (
	"time"
)

type persistMessage struct {
	Timestamp time.Time
	Rssi      int
	Mac       string
	Name      string
}
