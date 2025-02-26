package main

import (
	"math/rand"
	"time"
)

type Account struct {
	Id        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Number    int64     `json:"number"`
	Balance   uint64    `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Balance   uint64 `json:"balance"`
}

func NewAccount(firstName string, lastName string, balance uint64) *Account {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Number:    int64(rand.Intn(10000)),
		Balance:   balance,
		CreatedAt: time.Now().UTC(),
	}
}
