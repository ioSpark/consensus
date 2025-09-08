package test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"consensus/app"
)

func init() {
	registerRepoTest("UserDeleteNonExistent", testUserDeleteNonExistent)
	registerRepoTest("UserCreateDuplicate", testUserCreateDuplicate)
	registerRepoTest("UserDeleteNonExistent", testUserDeleteNonExistent)
	registerRepoTest("UserNoLeakage", testUserNoLeakage)
	registerRepoTest("UserCRUD", testUserCRUD)
}

func testUserCRUD(t *testing.T, repo app.Repository) {
	if len(repo.Users()) != 0 {
		t.Fatalf("expected 0 tickets: got %d", len(repo.Users()))
	}

	// TODO: CreateUser is inconsistent with CreateTicket
	u1 := app.UserID("1")
	err := repo.CreateUser(u1)
	if err != nil {
		t.Fatalf("create user1 failed: %v", err)
	}
	u2 := app.UserID("2")
	err = repo.CreateUser(u2)
	if err != nil {
		t.Fatalf("create user2 failed: %v", err)
	}

	if len(repo.Users()) != 2 {
		t.Fatalf("expected 2 users: got %d", len(repo.Users()))
	}

	_, err = repo.User(string(u1))
	if err != nil {
		t.Fatalf("could not fetch created user %v", u1)
	}

	if !slices.Contains(repo.Users(), u1) {
		t.Errorf("Users() missing created users: %v", u1)
	}

	err = repo.DeleteUser(u1)
	if err != nil {
		t.Fatalf("deleting user failed: %v", err)
	}

	_, err = repo.User(string(u1))
	if err == nil {
		t.Fatal("expected non-existent user to fail")
	}
	if !errors.Is(err, app.ErrUserNotExist) {
		t.Fatalf("expected ErrUserNotExist, got %v", err)
	}
}

func testUserDeleteNonExistent(t *testing.T, repo app.Repository) {
	err := repo.DeleteUser(app.UserID("non-existent"))
	if err == nil {
		t.Fatal("expected non-existent user update to fail")
	}
	if !errors.Is(err, app.ErrUserNotExist) {
		t.Fatalf("expected ErrUserNotExist, got %v", err)
	}
}

func testUserCreateDuplicate(t *testing.T, repo app.Repository) {
	u := app.NewUser("duplicate")
	if err := repo.CreateUser(u); err != nil {
		t.Fatalf("first create user failed: %v", err)
	}

	if err := repo.CreateUser(u); err == nil {
		fmt.Println(repo.Users())
		t.Fatal("expected duplicate user error")
	} else if !errors.Is(err, app.ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

// TODO: Characterisation test - memory implementation leaks
func testUserNoLeakage(t *testing.T, repo app.Repository) {
	_ = repo.CreateUser("1")
	_ = repo.CreateUser("2")

	users := repo.Users()
	if len(users) != 2 {
		t.Fatalf("expected 2 users got %d", len(users))
	}

	users[0] = "3"

	refetch, err := repo.User("3")
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if refetch != "3" {
		t.Errorf("expected leakage %s, got %s", "3", refetch)
	}
}
