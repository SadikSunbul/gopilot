package main

import (
	"fmt"
	"log"

	"github.com/SadikSunbul/Gopilot/clients"
)

func main() {
	apiKey := "your-api-key-here"

	client, err := clients.NewGeminiClient(apiKey, "gemini-2.0-flash")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	response, err := client.Stream("Will AI take jobs from computer engineers?")
	if err != nil {
		log.Fatal(err)
	}

	// When the channel is closed, the loop automatically ends
	for chunk := range response {
		fmt.Print(chunk)
	}
}
