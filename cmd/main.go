package main

import (
	"echoAuth/config"
	"log"
)

func main() {

	if err := config.Init(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Loaded succssfully")
}
