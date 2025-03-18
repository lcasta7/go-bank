package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type HttpRequest interface {
	GetAccountNumber() int64
}

type TransferRequest struct {
	FromNumber int64  `json:"from_number"`
	ToNumber   int64  `json:"to_number"`
	Amount     uint64 `json:"amount"`
}

func (r *TransferRequest) GetAccountNumber() int64 {
	return r.FromNumber
}

type DeleteAccountRequest struct {
	AdminAccount int64 `json:"admin_account"`
}

func (r *DeleteAccountRequest) GetAccountNumber() int64 {
	return r.AdminAccount
}

type GetAccountRequest struct {
	Number int64 `json:"number"`
}

func (r *GetAccountRequest) GetAccountNumber() int64 {
	return r.Number
}

type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Number int64  `json:"number"`
	Token  string `json:"token"`
}

type Account struct {
	Id                int       `json:"id"`
	FirstName         string    `json:"firstName"`
	LastName          string    `json:"lastName"`
	Number            int64     `json:"number"`
	EncryptedPassword string    `json:"-"`
	Balance           uint64    `json:"balance"`
	Role              string    `json:"role"`
	CreatedAt         time.Time `json:"createdAt"`
}

func (a *Account) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword), []byte(password))
}

type CreateAccountRequest struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	Balance      uint64 `json:"balance"`
	AdminAccount int64  `json:"admin_account"`
}

func (r *CreateAccountRequest) GetAccountNumber() int64 {
	return r.AdminAccount
}

func NewAccount(firstName string, lastName string, password string, role string, balance uint64) (*Account, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Account{
		FirstName:         firstName,
		LastName:          lastName,
		Number:            int64(rand.Intn(10000)),
		EncryptedPassword: string(encpw),
		Balance:           balance,
		Role:              role,
		CreatedAt:         time.Now().UTC(),
	}, nil
}
