package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	acc, err := NewAccount("first", "test", "password", "user", 100)

	assert.Nil(t, err)
	assert.Equal(t, acc.FirstName, "first")
	assert.Equal(t, acc.LastName, "test")
	assert.Equal(t, acc.Role, "user")
	assert.Equal(t, acc.Balance, uint64(100))
}

func TestValidatePassword(t *testing.T) {
	acc, _ := NewAccount("first", "test", "secret123", "user", 100)

	err := acc.ValidatePassword("secret123")
	assert.Nil(t, err)

	err = acc.ValidatePassword("wrong_password")
	assert.NotNil(t, err)
}

func TestGetAccountNumber(t *testing.T) {
	transferReq := TransferRequest{
		FromNumber: 1234,
	}
	assert.Equal(t, transferReq.FromNumber, int64(1234))

	deleteAccReq := DeleteAccountRequest{
		AdminAccount: 987,
	}
	assert.Equal(t, deleteAccReq.AdminAccount, int64(987))

	getAccReq := GetAccountRequest{
		Number: 876,
	}
	assert.Equal(t, getAccReq.Number, int64(876))

	createAccReq := CreateAccountRequest{
		AdminAccount: 1337,
	}
	assert.Equal(t, createAccReq.AdminAccount, int64(1337))
}

func TestNewAccountUniqueness(t *testing.T) {
	acc1, _ := NewAccount("first", "test", "password", "user", 100)
	acc2, _ := NewAccount("second", "test", "password", "user", 100)

	assert.NotEqual(t, acc1.Number, acc2.Number)
	assert.NotEqual(t, acc1.FirstName, acc2.FirstName)
}

func TestPasswordOmission(t *testing.T) {
	acc, _ := NewAccount("first", "last", "secretpass123", "user", 100)

	jsonAcc, err := json.Marshal(acc)
	assert.Nil(t, err)

	jsonStr := string(jsonAcc)

	assert.Contains(t, jsonStr, "first")
	assert.Contains(t, jsonStr, "last")

	assert.NotContains(t, jsonStr, "secretpass123")
}
