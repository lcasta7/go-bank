package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateNewAccount(t *testing.T) {
	acc, err := NewAccount("first", "test", "password", 100)

	assert.Nil(t, err)
	assert.Equal(t, acc.FirstName, "first")
	assert.Equal(t, acc.LastName, "test")
	assert.Equal(t, acc.Balance, uint64(100))
}
