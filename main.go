package main

import (
	"log"
	"sync"
)

func main() {

	dbFile := "suppedas.db"

	persistChannel := make(chan persistMessage)
	var wg sync.WaitGroup
	wg.Add(1)
	go bluetoothCtl(&wg, persistChannel)
	go databaseWriter(dbFile, persistChannel)
	log.Printf("Suppedas started, using db: %s\n", dbFile)
	wg.Wait()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
