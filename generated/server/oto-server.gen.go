// Code generated by oto; DO NOT EDIT.

package server

import (
	"context"
	"net/http"

	"github.com/pacedotdev/oto/otohttp"
)

// TacoService contains all knowledge around consumption of tacos
type TacoService interface {

	// EatTaco handles keeping track of eating tacos
	EatTaco(context.Context, EatTacoRequest) (*EatTacoResponse, error)
}

type tacoServiceServer struct {
	server      *otohttp.Server
	tacoService TacoService
}

// Register adds the TacoService to the otohttp.Server.
func RegisterTacoService(server *otohttp.Server, tacoService TacoService) {
	handler := &tacoServiceServer{
		server:      server,
		tacoService: tacoService,
	}
	server.Register("TacoService", "EatTaco", handler.handleEatTaco)
}

func (s *tacoServiceServer) handleEatTaco(w http.ResponseWriter, r *http.Request) {
	var request EatTacoRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.tacoService.EatTaco(r.Context(), request)
	if err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

// EatTacoRequest is the request for TacoService.EatTaco.
type EatTacoRequest struct {
	// Name is your name
	Name string `json:"name"`
	// All of the Tacos you have consumed 🌮
	Tacos []string `json:"tacos"`
}

// EatTacoResponse is the response for TacoService.EatTaco.
type EatTacoResponse struct {
	// TacoConsumptionStatus is your current taco consumption status
	TacoConsumptionStatus string `json:"tacoConsumptionStatus"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty"`
}
