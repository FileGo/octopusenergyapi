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

	client, err := octopusenergyapi.NewClient("https://api.octopus.energy/v1/", "{API_KEY}", http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	mpoint, err := client.GetMeterPoint(mpan)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("MPAN: %s\nProfile class: %d (%s)\n", mpoint.MPAN, mpoint.ProfileClass, octopusenergyapi.PCs[mpoint.ProfileClass])
	fmt.Printf("GSP: %s (%s)\n", mpoint.GSP.GSPGroupID, mpoint.GSP.Name)
}
