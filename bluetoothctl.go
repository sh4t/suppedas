package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	expect "github.com/google/goexpect"
)

func bluetoothCtl(wg *sync.WaitGroup, persistChannel chan persistMessage, recordResolution uint32) {
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
	log.Printf("Connected to bluetoothctl, RSSI record resolution is %d seconds.", recordResolution)

	// [NEW] Device E4:7D:BD:55:6B:22 [TV] Samsung 7 Series (55)
	var newEntryRegex = regexp.MustCompile(`^\[.*NEW.*\] Device [A-Z,0-9,:]* .*\n`)
	// [CHG] Device 5D:DD:24:01:FA:A1 RSSI: -44
	var rssiChangedRegex = regexp.MustCompile(`^\[.*CHG.*\] Device [A-Z,0-9,:]* RSSI: -[0-9]{1,3}\n`)

	var cases []expect.Caser
	cases = append(cases, &expect.Case{R: rssiChangedRegex}) // index 1
	cases = append(cases, &expect.Case{R: newEntryRegex})    // index 0

	// emit warning if no matches in secondsToWarn
	secondsToWarn := 60.0
	lastMatch := time.Now()

	// used to throttle rssi messsages
	var rssiMessages map[string]*persistMessage
	rssiMessages = make(map[string]*persistMessage)
	removeChannel := make(chan string)

	for {
		_, match, index, _ := e.ExpectSwitchCase(cases, 10*time.Millisecond)

		// if an entry requested to be removed, do it so...
		select {
		case mac := <-removeChannel:
			entry := rssiMessages[mac]
			persistChannel <- *entry
			log.Printf("Persisting: %v", entry)
			delete(rssiMessages, mac)
		default:
		}

		// did the regex match?
		if len(match) > 0 {
			switch index {
			case 0:
				matchSplit := strings.Split(match[0], " ")
				mac := matchSplit[2]
				rssi, err := strconv.Atoi(strings.TrimSuffix(matchSplit[4], "\n"))
				if err != nil {
					log.Printf("Error reading rssi in message: %v\n", matchSplit)
					continue
				}
				entry := rssiMessages[mac]

				if entry != nil {
					if rssi > entry.Rssi {
						entry.Rssi = rssi
						entry.Timestamp = time.Now()
					}
				} else {
					rssiMessages[mac] = &persistMessage{Mac: mac, Rssi: rssi, Timestamp: time.Now()}
					go entryRemover(mac, removeChannel, recordResolution)
				}

				lastMatch = time.Now()
			case 1:
				matchSplit := strings.Split(match[0], " ")
				mac := matchSplit[2]
				name := strings.SplitN(match[0], " ", 4)[3]
				name = strings.TrimSuffix(name, "\n")
				message := persistMessage{Mac: mac, Name: name, Timestamp: time.Now()}
				persistChannel <- message
				lastMatch = time.Now()

			}
		} else {
			if time.Since(lastMatch).Seconds() > secondsToWarn {
				log.Printf("Warning, no bluetooth activity for %f seconds", secondsToWarn)
				lastMatch = time.Now()
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

func entryRemover(mac string, removeChannel chan string, timeoutSeconds uint32) {
	time.Sleep(time.Duration(timeoutSeconds) * time.Second)
	removeChannel <- mac
}
