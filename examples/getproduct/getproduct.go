package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/FileGo/octopusenergyapi"
)

func main() {
	productCode := "VAR-17-01-11"

	client, err := octopusenergyapi.NewClient("{API_KEY}", http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	prod, err := client.GetProduct(productCode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s\n", prod.FullName)
	fmt.Printf("Is green: %t\n", prod.IsGreen)
	fmt.Printf("Available from %s to %s\n", prod.AvailableFrom, prod.AvailableTo)
}
