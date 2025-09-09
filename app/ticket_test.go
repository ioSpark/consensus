package app_test

import (
	"fmt"
	"testing"

	"consensus/app"
	"consensus/repo/memory"
)

// TODO: These are the same functions used in the repo tests
func createUser(t *testing.T, repo app.Repository, name string) *app.UserID {
	t.Helper()

	u := app.NewUser(name)
	err := repo.CreateUser(u)
	if err != nil {
		t.Fatalf("create user '%s' failed: %v", name, err)
	}

	return &u
}

func createTicket(
	t *testing.T,
	repo app.Repository,
	name string,
	user app.UserID,
) *app.Ticket {
	t.Helper()

	ticket, err := repo.CreateTicket(
		app.NewTicket(name, "http://whatever"+name, user),
	)
	if err != nil {
		t.Errorf("CreateTicket failed: %v", err)
	}

	return &ticket
}

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

	repo := memory.NewRepository()
	reporter := createUser(t, repo, "reporter")

	for iTicket, test := range tests {
		ticket := createTicket(
			t,
			repo,
			fmt.Sprintf("test-ticket-%d", iTicket),
			*reporter,
		)

		for iInput, input := range test.Input {
			u := createUser(t, repo, fmt.Sprintf("user-%d-%d", iTicket, iInput))

			_, err := repo.Vote(ticket.ID, *u, input)
			if err != nil {
				t.Errorf("couldn't add point %d: %s", input, err)
			}
		}

		refreshed, err := repo.Ticket(ticket.ID)
		if err != nil {
			t.Fatalf("get ticket failed after pointing: %v", err)
		}

		result := refreshed.Mode()
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
	repo := memory.NewRepository()
	u := createUser(t, repo, "test")

	ticket := createTicket(t, repo, "test-ticket", *u)

	if ticket.Voted(*u) != false {
		t.Error("voted is True, but we haven't voted")
	}

	refreshed, err := repo.Vote(ticket.ID, *u, 2)
	if err != nil {
		t.Errorf("unexpected error while pointing: %s", err)
	}
	if refreshed.Voted(*u) != true {
		t.Error("voted is False, but we have voted")
	}
}
