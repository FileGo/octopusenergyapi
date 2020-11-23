package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/FileGo/octopusenergyapi"
)

func main() {
	client, err := octopusenergyapi.NewClient("https://api.octopus.energy/v1/", "{API_KEY}", *http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	products, err := client.ListProducts()
	if err != nil {
		log.Fatal(err)
	}

	for _, product := range products {
		fmt.Printf("[%s] %s\n", product.Code, product.DisplayName)
	}

	fmt.Printf("Number of products: %d\n", len(products))

}
