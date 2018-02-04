package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	expect "github.com/google/goexpect"
)

func bluetoothCtl(wg *sync.WaitGroup, persistChannel chan persistMessage) {
	defer wg.Done()

	e, _, err := expect.Spawn("bluetoothctl", -1)
	check(err)

	var startupRegex = regexp.MustCompile(`Agent registered\n`)
	_, _, err = e.Expect(startupRegex, 5*time.Second)
	check(err)

	err = e.Send("power on\n")
	check(err)
	err = e.Send("set-scan-filter-rssi -100\n")
	check(err)
	err = e.Send("scan on\n")
	check(err)
	err = e.Send("remove *\r\n")
	check(err)
	log.Printf("Connected to bluetoothctl, scanning started.")

	// [NEW] Device E4:7D:BD:55:6B:22 [TV] Samsung 7 Series (55)
	var newEntryRegex = regexp.MustCompile(`^\[.*NEW.*\] Device [A-Z,0-9,:]* .*\n`)
	// [CHG] Device 5D:DD:24:01:FA:A1 RSSI: -44
	var rssiChangedRegex = regexp.MustCompile(`^\[.*CHG.*\] Device [A-Z,0-9,:]* RSSI: -[0-9]{1,3}\n`)

	var cases []expect.Caser
	cases = append(cases, &expect.Case{rssiChangedRegex, "", nil, 0}) // index 1
	cases = append(cases, &expect.Case{newEntryRegex, "", nil, 0})    // index 0

	for {
		_, match, index, _ := e.ExpectSwitchCase(cases, 10*time.Millisecond)
		if len(match) > 0 {
			switch index {
			case 0:
				matchSplit := strings.Split(match[0], " ")
				mac := matchSplit[2]
				rssi := strings.TrimSuffix(matchSplit[4], "\n")
				message := persistMessage{Mac: mac, Rssi: rssi, Timestamp: time.Now()}
				persistChannel <- message
			case 1:
				matchSplit := strings.Split(match[0], " ")
				mac := matchSplit[2]
				name := strings.SplitN(match[0], " ", 4)[3]
				name = strings.TrimSuffix(name, "\n")
				message := persistMessage{Mac: mac, Name: name, Timestamp: time.Now()}
				persistChannel <- message
			}
		}
	}
}

func checkExpectErr(out string, match []string, e error) {
	if e != nil {
		fmt.Printf("Gexpect error: %s\n", e)
		fmt.Printf("Gexpect out: %s\n", out)
		fmt.Printf("Gexpect match: %v\n", match)
		panic(e)
	}
}
