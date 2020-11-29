package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/FileGo/octopusenergyapi"
)

func main() {
	postcode := "SW1A 1AA"

	client, err := octopusenergyapi.NewClient("{API_KEY}", http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	gsp, err := client.GetGridSupplyPoint(postcode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Postcode: %s\n", postcode)
	fmt.Printf("Name: %s\nOperator: %s\nPhone number: %s\nParticipant ID: %s\nGroup ID: %s\n",
		gsp.Name, gsp.Operator, gsp.PhoneNumber, gsp.ParticipantID, gsp.GSPGroupID)
}
