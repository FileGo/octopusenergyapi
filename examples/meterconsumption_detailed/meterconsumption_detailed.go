package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FileGo/octopusenergyapi"
)

func main() {
	// Meter point's MPAN
	mpan := "1234567890"
	// Meter's serial number
	serialno := "1234567890"

	client, err := octopusenergyapi.NewClient("{API_KEY}", http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve data for last 7 days
	options := octopusenergyapi.ConsumptionOption{
		From:     time.Now().AddDate(0, 0, -7),
		To:       time.Now(),
		PageSize: 25000,
	}

	cons, err := client.GetMeterConsumption(mpan, serialno, options)
	if err != nil {
		log.Fatal(err)
	}

	total := float32(0.0)
	for i, line := range cons {
		total += line.Value
		fmt.Printf("[%d] From: %s To: %s Value: %1.3f\n", i, line.IntervalStart, line.IntervalEnd, line.Value)
	}

	fmt.Printf("Total consumption: %1.3f\n", total)
}
