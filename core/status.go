package core
type Status struct {
	EndCards   []Card `json:"endcards"`
	StackCards []Card `json:"stackcards"`
	HandCards  []Card `json:"handcards"`
}
