package main

import "math/rand"

type Account struct {
	Id        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Number    int64  `json:"number"`
	Balance   int64  `json:"balance"`
}

func NewAccount(firstName string, lastName string) *Account {
	return &Account{
		Id:        rand.Intn(10000),
		FirstName: firstName,
		LastName:  lastName,
		Number:    int64(rand.Intn(10000)),
		Balance:   int64(rand.Intn(900)),
	}
}
