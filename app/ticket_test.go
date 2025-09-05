package app_test

import (
	"fmt"
	"testing"

	"consensus/app"
)

type mode struct {
	Input  []int
	Output []app.Point
}

func TestMode(t *testing.T) {
	// Maybe worth splitting this up into separate tests
	tests := []mode{
		{
			Input:  []int{1, 2, 3, 3, 5},
			Output: []app.Point{3},
		},
		{
			Input:  []int{5, 5, 5, 5, 5},
			Output: []app.Point{5},
		},
		{
			Input:  []int{1, 1, 2, 2, 5, 5},
			Output: []app.Point{1, 2, 5},
		},
		{
			Input:  []int{1, 2, 3, 5, 8, 13},
			Output: []app.Point{1, 2, 3, 5, 8, 13},
		},
		{
			Input:  []int{1, 1, 2, 3, 5, 5},
			Output: []app.Point{1, 5},
		},
		{
			Input:  []int{1, 2, 3, 5, 8, 13},
			Output: []app.Point{1, 2, 3, 5, 8, 13},
		},
		// If it's ever called without points being added to the ticket
		{
			Input:  []int{},
			Output: nil,
		},
	}

	for _, test := range tests {
		dummyUser := app.NewUser("reporter")
		ticket := app.NewTicket("a", "b", dummyUser)

		for i, input := range test.Input {
			u := app.NewUser(fmt.Sprintf("user-%d", i))

			err := ticket.Vote(u, input)
			if err != nil {
				t.Errorf("couldn't add point %d: %s", input, err)
			}
		}

		result := ticket.Mode()
		if len(test.Output) != len(result) {
			t.Errorf("mismatched results, expected %v, got %v", test.Output, result)
		}
		for i, v := range result {
			p := app.Point(test.Output[i])
			if p != v {
				t.Errorf("mismatched results, expected %v, got %v", test.Output, result)
				return
			}
		}
	}
}

func TestVoted(t *testing.T) {
	u := app.NewUser("test")
	ticket := app.NewTicket("a", "b", u)

	if ticket.Voted(u) != false {
		t.Error("voted is True, but we haven't voted")
	}

	err := ticket.Vote(u, 2)
	if err != nil {
		t.Errorf("unexpected error while pointing: %s", err)
	}
	if ticket.Voted(u) != true {
		t.Error("voted is False, but we have voted")
	}
}
