package definitions

// TacoService contains all knowledge around consumption of tacos
type TacoService interface {

	// EatTaco handles keeping track of eating tacos
	EatTaco(request EatTacoRequest) EatTacoResponse
}

// EatTacoRequest is the request for TacoService.EatTaco.
type EatTacoRequest struct {
	// Name is your name
	Name string
	// All of the Tacos you have consumed 🌮
	Tacos []string
}

// EatTacoResponse is the response for TacoService.EatTaco.
type EatTacoResponse struct {
	// Your current taco consumption status
	TacoConsumptionStatus string
}
