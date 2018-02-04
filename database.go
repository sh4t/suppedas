package main

import "fmt"

func databaseWriter(persistChannel chan persistMessage) {
	for {
		entry := <-persistChannel
		fmt.Printf("%v ", entry)
	}
}
