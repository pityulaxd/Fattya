package core

import (
	"math/rand"
	"time"
)

// Deck holds the cards in the deck to be shuffled
type Deck struct {
	Cards []Card
}

// Card holds the card suits and types in the deck
type Card struct {
	Type int    `json:"type"`
	Suit string `json:"suit"`
	Ask  int    `json:"ask"`
}

//SetCard set a card
type SetCard interface {
	MakeCard(i int, s string)
}

//MakeCard make those cards boi
func (c Card) MakeCard(i int, s string) {
	c.Suit = s
	c.Type = i
}

// New creates a deck of cards to be used
func New(d int) []Card {
	var tempHand []Card
	types := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}

	// Valid suits include Heart, Diamond, Club & Spade
	suits := []string{"h", "d", "c", "s"}

	// Loop over each type and suit appending to the deck
	for j := 0; j < d; j++ {
		for i := 0; i < len(types); i++ {
			for n := 0; n < len(suits); n++ {
				card := Card{
					Type: types[i],
					Suit: suits[n],
				}
				tempHand = append(tempHand, card)
			}
		}
	}
	Shuffle(tempHand)
	return tempHand
}

//If in need to convert
//func MakeCard(deck Deck, card map[string]int){
//
//}

// Shuffle the deck
func Shuffle(d []Card) []Card {
	for i := 1; i < len(d); i++ {
		r := rand.Intn(i + 1)

		if i != r {
			d[r], d[i] = d[i], d[r]
		}
	}
	return d
}

var i int

// Draw card from the deck
func Draw(deck []Card) Card {
	if i < len(deck) {
		var card = Card{
			Type: deck[i].Type,
			Suit: deck[i].Suit}
		i++
		return card
	}
	panic("End of deck")
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
