package main

import (
	"flag"
	"log"
	"os"
	"sync"
)

func main() {

	var dbFile string
	var locationName string
	cmdFlags := flag.NewFlagSet("suppedasFlags", flag.ExitOnError)
	cmdFlags.StringVar(&locationName, "l", "", "Name of the current position")
	cmdFlags.StringVar(&dbFile, "d", "suppedas.db", "Output database")

	cmdFlags.Parse(os.Args[1:])

	if locationName == "" {
		cmdFlags.Usage()
		os.Exit(2)
	}

	persistChannel := make(chan persistMessage)
	var wg sync.WaitGroup
	wg.Add(1)
	go bluetoothCtl(&wg, persistChannel)
	go databaseWriter(dbFile, locationName, persistChannel)
	log.Printf("Suppedas started, using db: %s\n", dbFile)
	wg.Wait()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
