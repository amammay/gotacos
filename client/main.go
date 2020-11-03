package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/amammay/otorun/generated/client"
)

func main() {

	// if we are running on cloud run we will leverage runsd for service discovery
	tacoHost := "http://tacoserver/oto/"
	if !metadata.OnGCE() {
		tacoHost = "http://localhost:8080/oto/"
	}

	// create a new client from our auto generated client
	c := client.New(tacoHost)
	c.Debug = func(s string) {
		fmt.Println(s)
	}
	// create our taco service
	tacoService := client.NewTacoService(c)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// consume some tacos
		tacoResponse, err := tacoService.EatTaco(context.Background(), client.EatTacoRequest{
			Name:  "Sammy Sosa",
			Tacos: []string{"Chicken", "Chorizo"},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, tacoResponse.TacoConsumptionStatus)
	})

	// cloud run sets a port env variable that we should respect
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))

}
