package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/FileGo/octopusenergyapi"
)

func main() {
	// Meter point's MPAN
	mpan := "1234567890"
	// Meter's serial number
	serialno := "1234567890"

	client, err := octopusenergyapi.NewClient("API_KEY}", http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	cons, err := client.GetMeterConsumption(mpan, serialno, octopusenergyapi.ConsumptionOption{})
	if err != nil {
		log.Fatal(err)
	}

	for i, line := range cons {
		fmt.Printf("[%d] From: %s To: %s Value: %1.3f\n", i, line.IntervalStart, line.IntervalEnd, line.Value)
	}

}
