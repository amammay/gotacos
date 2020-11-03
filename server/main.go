package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/amammay/otorun/generated/server"
	"github.com/pacedotdev/oto/otohttp"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}

func run() error {

	// creates new oto base http server
	oto := otohttp.NewServer()

	// register our taco service to the oto server
	server.RegisterTacoService(oto, &tacoService{})

	mux := http.NewServeMux()

	// map the path prefix /oto/ to the oto server and our custom middleware
	mux.Handle("/oto/", userAgentMiddleware(oto))

	// cloud run sets a port env variable that we should respect
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// create our http server and bind our handle to our mux
	httpServer := &http.Server{
		Handler:      mux,
		Addr:         ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Printf("Starting server on %s", httpServer.Addr)

	return httpServer.ListenAndServe()

}

// userAgentMiddleware logs the user agent middleware for all requests that come into oto
func userAgentMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if agent := r.UserAgent(); agent != "" {
			log.Printf("info: request from %s", agent)
		}

		// send the request into oto
		h.ServeHTTP(w, r)
	}

}

// tacoService implementation
type tacoService struct {
	// TODO add database
}

// EatTaco will tell you how many tacos you have consumed
func (t *tacoService) EatTaco(ctx context.Context, request server.EatTacoRequest) (*server.EatTacoResponse, error) {
	if request.Name == "" {
		return nil, errors.New("EatTacoRequest.Name is required")
	}
	if len(request.Tacos) <= 0 {
		return nil, errors.New("EatTacoRequest.Tacos is required")
	}

	return &server.EatTacoResponse{TacoConsumptionStatus: fmt.Sprintf("%s has consumed %d tacos", request.Name, len(request.Tacos))}, nil
}
