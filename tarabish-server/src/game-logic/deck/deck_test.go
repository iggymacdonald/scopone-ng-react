package deck

import (
	"fmt"
	"testing"
)

func TestFindFirstCard(t *testing.T) {
	deck := New()
	searchI := 0
	i, found := Find(deck, deck[searchI])
	if !found {
		t.Errorf("Card %v should be found", deck[0])
	}
	if TypeSuit(deck[i]) != TypeSuit(deck[searchI]) {
		t.Errorf("Card %v should be equal to card %v", deck[i], deck[searchI])
	}
}

func TestFindLastCard(t *testing.T) {
	deck := New()
	searchI := len(deck) - 1
	i, found := Find(deck, deck[searchI])
	if !found {
		t.Errorf("Card %v should be found", deck[0])
	}
	if TypeSuit(deck[i]) != TypeSuit(deck[searchI]) {
		t.Errorf("Card %v should be equal to card %v", deck[i], deck[searchI])
	}
}

func TestFindCardInTheMiddle(t *testing.T) {
	deck := New()
	searchI := 10
	i, found := Find(deck, deck[searchI])
	if !found {
		t.Errorf("Card %v should be found", deck[0])
	}
	if TypeSuit(deck[i]) != TypeSuit(deck[searchI]) {
		t.Errorf("Card %v should be equal to card %v", deck[i], deck[searchI])
	}
}

func TestRemoveFirstCard(t *testing.T) {
	deck := New()
	cardsAfter := RemoveCard(deck, deck[0])
	if len(cardsAfter) != len(deck)-1 {
		t.Errorf("After removing one card the cards should be %v but instead are %v", len(deck)-1, len(cardsAfter))
	}
	for i := range cardsAfter {
		if TypeSuit(cardsAfter[i]) == TypeSuit(deck[0]) {
			t.Errorf("Card %v should be removed but it is still there at index %v", TypeSuit(deck[0]), i)
		}
	}
}

func TestRemoveLastCard(t *testing.T) {
	deck := New()
	cardsAfter := RemoveCard(deck, deck[len(deck)-1])
	if len(cardsAfter) != len(deck)-1 {
		t.Errorf("After removing one card the cards should be %v but instead are %v", len(deck)-1, len(cardsAfter))
	}
	for i := range cardsAfter {
		if TypeSuit(cardsAfter[i]) == TypeSuit(deck[len(deck)-1]) {
			t.Errorf("Card %v should be removed but it is still there at index %v", TypeSuit(deck[len(deck)-1]), i)
		}
	}
}

func TestRemoveTwoCardsInTheMiddle(t *testing.T) {
	deck := New()
	iRemove := 10
	firstCardToRemove := deck[iRemove]
	secondCardToRemove := deck[iRemove+1]
	cardsAfter := RemoveCard(deck, firstCardToRemove)
	cardsAfter = RemoveCard(cardsAfter, secondCardToRemove)
	if len(cardsAfter) != len(deck)-2 {
		t.Errorf("After removing one card the cards should be %v but instead are %v", len(deck)-2, len(cardsAfter))
	}
	for i := range cardsAfter {
		if TypeSuit(cardsAfter[i]) == TypeSuit(firstCardToRemove) {
			t.Errorf("Card %v should be removed but it is still there at index %v", TypeSuit(firstCardToRemove), i)
		}
		if TypeSuit(cardsAfter[i]) == TypeSuit(secondCardToRemove) {
			t.Errorf("Card %v should be removed but it is still there at index %v", TypeSuit(secondCardToRemove), i)
		}
	}
}

func TestRemoveCards(t *testing.T) {
	deck := New()
	iRemove1 := 10
	iRemove2 := 20
	cardsToRemove := []Card{deck[iRemove1], deck[iRemove2]}
	cardsAfter := RemoveCards(deck, cardsToRemove)
	if len(cardsAfter) != len(deck)-2 {
		t.Errorf("After removing one card the cards should be %v but instead are %v", len(deck)-2, len(cardsAfter))
	}
	for i := range cardsAfter {
		if TypeSuit(cardsAfter[i]) == TypeSuit(cardsToRemove[0]) {
			t.Errorf("Card %v should be removed but it is still there at index %v", TypeSuit(cardsToRemove[0]), i)
		}
		if TypeSuit(cardsAfter[i]) == TypeSuit(cardsToRemove[1]) {
			t.Errorf("Card %v should be removed but it is still there at index %v", TypeSuit(cardsToRemove[1]), i)
		}
	}
}

func TestNew(t *testing.T) {
	expectedNumberOfCards := 36
	deck := New()
	if len(deck) != expectedNumberOfCards {
		t.Errorf("Expected %v cards but found %v", expectedNumberOfCards, len(deck))
	}
	for i := 0; i < len(deck); i++ {
		fmt.Printf("%s of %s :: ", deck[i].Type, deck[i].Suit)
	}
	// for card := deck {
	// 	fmt.Println(card)
	// }
	// tests := []struct {
	// 	name     string
	// 	wantDeck Deck
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		if gotDeck := New(); !reflect.DeepEqual(gotDeck, tt.wantDeck) {
	// 			t.Errorf("New() = %v, want %v", gotDeck, tt.wantDeck)
	// 		}
	// 	})
	// }
}

// func TestShuffle(t *testing.T) {
// 	type args struct {
// 		d Deck
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want Deck
// 	}{
// 		// TODO: Add test cases.
// 		{"Full",
// 			args{New()},
// 			New()},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := Shuffle(tt.args.d); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Shuffle() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
