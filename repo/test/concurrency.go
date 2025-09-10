package test

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"consensus/app"
)

func init() {
	registerRepoTest("ConcurrentVoting", testConcurrentVoting)
	registerRepoTest("ConcurrentTicketCRUD", testConcurrentTicketCRUD)
}

func testConcurrentTicketCRUD(t *testing.T, repo app.Repository) {
	const userCount = 16
	// Seed users
	users := make([]app.UserID, userCount)
	for i := range userCount {
		u := createUser(t, repo, fmt.Sprintf("user-%d", i))
		users[i] = *u
	}

	// Seed a single user, so that we can pick a random ticket
	_ = createTicket(t, repo, "seed", users[0])

	workers := runtime.NumCPU() * 2
	// Perhaps worth making this configurable (somehow). Or perhaps exposing a more
	// intense version that is opt-in, since it takes a while.
	ops := 256
	var wg sync.WaitGroup
	start := make(chan struct{})

	wg.Add(workers)
	for w := range workers {
		go func(ID int) {
			defer wg.Done()
			<-start

			for i := range ops {
				rngUser := users[rand.IntN(userCount)]

				switch rand.IntN(4) {
				case 0: // Create ticket
					ticket := app.NewTicket(
						fmt.Sprintf("ticket-w%d-%d", ID, i),
						fmt.Sprintf("whatever-w%d-%d", ID, i),
						rngUser,
					)
					_, err := repo.CreateTicket(ticket)
					if err != nil {
						t.Errorf("could not create ticket %s: %v", ticket.Name, err)
					}
				case 1: // Update ticket
					tickets := repo.Tickets()
					rngTicket := tickets[rand.IntN(len(tickets))]

					rngTicket.Name = fmt.Sprintf("updated by w-%d-%d", ID, i)
					_, err := repo.UpdateTicket(rngTicket)
					if err != nil {
						t.Errorf("could not update ticket: %v", err)
					}
				case 2: // Vote
					tickets := repo.Tickets()
					rngTicket := tickets[rand.IntN(len(tickets))]

					rngVote := app.PointValues[rand.IntN(len(app.PointValues))]
					_, err := repo.Vote(rngTicket.ID, rngUser, int(rngVote))
					if err != nil {
						t.Errorf("could not vote on ticket: %v", err)
					}
				case 3: // Try to reveal
					tickets := repo.Tickets()
					rngTicket := tickets[rand.IntN(len(tickets))]

					if len(rngTicket.Votes) == 0 {
						continue
					}

					_, err := repo.Reveal(rngTicket.ID, rngTicket.RaisedBy)
					if err != nil {
						t.Errorf("could not reveal ticket: %v", err)
					}
				}
			}
		}(w)
	}

	close(start)
	wg.Wait()

	// Invariant checks
	seenIDs := make(map[int]struct{})
	for _, ticket := range repo.Tickets() {
		if _, dup := seenIDs[ticket.ID]; dup {
			t.Errorf("duplicate ticket ID: %d", ticket.ID)
		}
		seenIDs[ticket.ID] = struct{}{}

		// Verify name/link uniqueness
		for _, other := range repo.Tickets() {
			if ticket.ID == other.ID {
				continue
			}
			if ticket.Name == other.Name {
				t.Errorf("duplicate ticket name: %s", ticket.Name)
			}
			if ticket.Link == other.Link {
				t.Errorf("duplicate ticket link: %s", ticket.Link)
			}
		}

		if ticket.Revealed && len(ticket.Votes) == 0 {
			t.Errorf("revealed ticket %d has zero votes", ticket.ID)
		}
	}
}

func testConcurrentVoting(t *testing.T, repo app.Repository) {
	reporter := createUser(t, repo, "reporter")
	ticket := createTicket(t, repo, "ticket", *reporter)

	const workers = 256
	var userSequence atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})

	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			<-start

			user := createUser(t, repo, fmt.Sprintf("v-%d", userSequence.Add(1)))

			ticket, err := repo.Ticket(ticket.ID)
			if err != nil {
				t.Errorf("fetch ticket failed: %v", err)
				return
			}

			// Vote random value
			val := app.PointValues[int(userSequence.Load())%len(app.PointValues)]
			_, err = repo.Vote(ticket.ID, *user, int(val))
			if err != nil {
				t.Errorf("vote failed: %v", err)
				return
			}
		}()
	}

	close(start)
	wg.Wait()

	final, err := repo.Ticket(ticket.ID)
	if err != nil {
		t.Fatalf("final ticket lookup failed: %v", err)
	}

	if len(final.Votes) != workers {
		t.Errorf("expected %d votes, got %d", workers, len(final.Votes))
	}
	users := repo.Users()
	if len(users) != workers+1 {
		t.Errorf("expected %d users, got %d", workers, len(users))
	}
}
