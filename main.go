package main

import "sync"

func main() {

	persistChannel := make(chan persistMessage)
	var wg sync.WaitGroup
	wg.Add(1)
	go bluetoothCtl(&wg, persistChannel)
	go databaseWriter(persistChannel)
	wg.Wait()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
