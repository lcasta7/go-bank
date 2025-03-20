package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
	store, _ := NewPostgressStore()
	store.Init()

	accountNumber := int64(1337)
	account := &Account{
		FirstName:         "Test",
		LastName:          "CreateAccount",
		Number:            accountNumber,
		EncryptedPassword: "secret123",
		Balance:           1000,
		Role:              "user",
		CreatedAt:         time.Now(),
	}

	store.CreateAccount(account)

	storedAccount, _ := store.GetAccountByNumber(accountNumber)
	assert.Equal(t, account.FirstName, storedAccount.FirstName)
	assert.Equal(t, account.LastName, storedAccount.LastName)
	assert.Equal(t, account.Number, storedAccount.Number)
	assert.Equal(t, account.Balance, storedAccount.Balance)
	assert.Equal(t, account.Role, storedAccount.Role)

	store.DeleteAccount(account.Id)

	_, err := store.GetAccountByNumber(accountNumber)
	assert.Error(t, err)
}

func TestGetAccounts(t *testing.T) {

	store, _ := NewPostgressStore()
	store.Init()

	firstAccountNumber := int64(1337)
	firstAccount := &Account{
		FirstName:         "Test",
		LastName:          "GetAccountByIdFirst",
		Number:            firstAccountNumber,
		EncryptedPassword: "secret123",
		Balance:           1000,
		Role:              "user",
		CreatedAt:         time.Now(),
	}
	secondAccountNumber := int64(1338)
	secondAccount := &Account{
		FirstName:         "Test",
		LastName:          "GetAccountByIdSecond",
		Number:            secondAccountNumber,
		EncryptedPassword: "secret123",
		Balance:           1000,
		Role:              "user",
		CreatedAt:         time.Now(),
	}

	store.CreateAccount(firstAccount)
	store.CreateAccount(secondAccount)

	accounts, _ := store.GetAccounts()

	var foundFirstAccount *Account
	for _, acc := range accounts {
		if acc.Number == firstAccountNumber {
			foundFirstAccount = acc
			break
		}
	}

	var foundSecondAccount *Account
	for _, acc := range accounts {
		if acc.Number == secondAccountNumber {
			foundSecondAccount = acc
			break
		}
	}

	assert.Equal(t, firstAccount.Number, foundFirstAccount.Number)
	assert.Equal(t, firstAccount.LastName, foundFirstAccount.LastName)
	assert.Equal(t, secondAccount.Number, foundSecondAccount.Number)
	assert.Equal(t, secondAccount.LastName, foundSecondAccount.LastName)

	store.DeleteAccount(firstAccount.Id)
	store.DeleteAccount(secondAccount.Id)
}

func TestTransferMoney(t *testing.T) {
	store, _ := NewPostgressStore()
	store.Init()

	fromAccountNumber := int64(1337)
	fromAccount := &Account{
		FirstName:         "Test",
		LastName:          "TransferMoneyFrom",
		Number:            fromAccountNumber,
		EncryptedPassword: "secret123",
		Balance:           1000,
		Role:              "user",
		CreatedAt:         time.Now(),
	}
	toAccountNumber := int64(1338)
	toAccount := &Account{
		FirstName:         "Test",
		LastName:          "TransferMoneyTo",
		Number:            toAccountNumber,
		EncryptedPassword: "secret123",
		Balance:           1,
		Role:              "user",
		CreatedAt:         time.Now(),
	}

	store.CreateAccount(fromAccount)
	store.CreateAccount(toAccount)

	store.TransferMoney(fromAccount, toAccount, 10)

	fromAccountUpdate, _ := store.GetAccountByNumber(fromAccount.Number)
	toAccountUpdate, _ := store.GetAccountByNumber(toAccount.Number)

	//check updated values
	assert.Equal(t, fromAccount.Balance-10, fromAccountUpdate.Balance)
	assert.Equal(t, toAccount.Balance+10, toAccountUpdate.Balance)

	store.DeleteAccount(fromAccount.Id)
	store.DeleteAccount(toAccount.Id)

}
