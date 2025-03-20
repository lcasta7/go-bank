package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

// this line is needed by the go generate command to generate gomock files
//go:generate mockgen -destination=mock_storage.go -package=main . Storage

func TestHandleGetAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStore := NewMockStorage(ctrl)
	server := NewApiServer(":3000", mockStore)

	//need to set a mock due to api.go:96 03-18-25
	mockStore.EXPECT().
		GetAccounts().
		Return([]*Account{
			{Id: 1, FirstName: "John", LastName: "Doe", Number: 1001, Role: "admin"},
		}, nil)

	requestBodyJson := `{"number": 1001}`
	req := httptest.NewRequest("GET", "/accounts", strings.NewReader(requestBodyJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-jwt-token", createTestJWT(t, 1001, "admin"))

	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/accounts", jwtAuthMiddleware(makeHttpHandleFunc(server.handleGetAccounts))).Methods("GET")

	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var accounts []*Account
	err := json.Unmarshal(recorder.Body.Bytes(), &accounts)
	require.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, int64(1001), accounts[0].Number)
	assert.Equal(t, "John", accounts[0].FirstName)
	assert.Equal(t, "Doe", accounts[0].LastName)
}

func TestHandleCreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStore := NewMockStorage(ctrl)
	server := NewApiServer(":3000", mockStore)

	mockStore.EXPECT().
		CreateAccount(gomock.Any()).
		Return(nil).
		Times(1)

	requestBodyJson := `{
		"firstName": "tars",
		"lastName": "robo",
		"password": "gobank",
		"balance": 20,
		"admin_account": 1337
	}`

	req := httptest.NewRequest("POST", "/account", strings.NewReader(requestBodyJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-jwt-token", createTestJWT(t, 1337, "admin"))

	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/account", jwtAuthMiddleware(makeHttpHandleFunc(server.handleCreateAccount))).Methods("POST")

	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	account := new(Account)
	err := json.Unmarshal(recorder.Body.Bytes(), &account)
	require.NoError(t, err)
	assert.Equal(t, account.Balance, uint64(20))
	assert.Equal(t, account.Role, "user")
	assert.Equal(t, account.FirstName, "tars")
	assert.Equal(t, account.LastName, "robo")
}

func TestHandleDeleteAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStore := NewMockStorage(ctrl)
	server := NewApiServer(":3000", mockStore)

	mockStore.EXPECT().
		DeleteAccount(gomock.Any()).
		Return(nil).
		Times(1)

	requestBodyJson := `{
	        "admin_account": 1337
	}`

	endpoint := fmt.Sprintf("/account/%d", 1337)
	req := httptest.NewRequest("DELETE", endpoint, strings.NewReader(requestBodyJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-jwt-token", createTestJWT(t, 1337, "admin"))

	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/account/{id}", jwtAuthMiddleware(makeHttpHandleFunc(server.handleDeleteAccount))).Methods("DELETE")

	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStore := NewMockStorage(ctrl)
	server := NewApiServer(":3000", mockStore)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	account := &Account{
		Id:                0,
		FirstName:         "John",
		LastName:          "Doe",
		Number:            9966,
		EncryptedPassword: string(hashedPassword),
		Balance:           1000,
		Role:              "user",
		CreatedAt:         time.Now(),
	}

	mockStore.EXPECT().
		GetAccountByNumber(account.Number).
		Return(account, nil).
		Times(1)

	requestBodyJson := `{
	        "number": 9966,
	        "password": "secret123"
	}`

	req := httptest.NewRequest("POST", "/login", strings.NewReader(requestBodyJson))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/login", makeHttpHandleFunc(server.handleLogin)).Methods("POST")
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var loginResponse *LoginResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	assert.Equal(t, int64(9966), loginResponse.Number)

}

func TestHandleGetAccountByNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStore := NewMockStorage(ctrl)
	server := NewApiServer(":3000", mockStore)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	account := &Account{
		Id:                0,
		FirstName:         "John",
		LastName:          "Doe",
		Number:            9966,
		EncryptedPassword: string(hashedPassword),
		Balance:           1000,
		Role:              "user",
		CreatedAt:         time.Now(),
	}

	mockStore.EXPECT().
		GetAccountByNumber(account.Number).
		Return(account, nil).
		Times(1)

	requestBodyJson := `{
	        "number": 9966
	}`

	req := httptest.NewRequest("GET", "/account", strings.NewReader(requestBodyJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-jwt-token", createTestJWT(t, 9966, "user"))

	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/account", jwtAuthMiddleware(makeHttpHandleFunc(server.handleGetAccountByNumber))).Methods("GET")
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var returnedAcc *Account
	err := json.Unmarshal(recorder.Body.Bytes(), &returnedAcc)
	require.NoError(t, err)

	assert.Equal(t, account.Number, returnedAcc.Number)
	assert.Equal(t, account.Balance, returnedAcc.Balance)
	assert.Equal(t, account.FirstName, returnedAcc.FirstName)
	assert.Equal(t, account.LastName, returnedAcc.LastName)

}

func TestHandleTransfer(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStore := NewMockStorage(ctrl)
	server := NewApiServer(":3000", mockStore)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	fromAccount := &Account{
		Id:                0,
		FirstName:         "John",
		LastName:          "Doe",
		Number:            9901,
		EncryptedPassword: string(hashedPassword),
		Balance:           1000,
		Role:              "user",
		CreatedAt:         time.Now(),
	}
	toAccount := &Account{
		Id:                0,
		FirstName:         "Doe",
		LastName:          "John",
		Number:            9902,
		EncryptedPassword: string(hashedPassword),
		Balance:           1,
		Role:              "user",
		CreatedAt:         time.Now(),
	}

	mockStore.EXPECT().
		GetAccountByNumber(fromAccount.Number).
		Return(fromAccount, nil).
		Times(2)
	mockStore.EXPECT().
		GetAccountByNumber(toAccount.Number).
		Return(toAccount, nil).
		Times(1)
	mockStore.EXPECT().
		TransferMoney(fromAccount, toAccount, uint64(500)).
		Return(nil).
		Times(1)

	requestBodyJson := `{
	   "from_number": 9901,
	   "to_number": 9902,
	   "amount": 500
	}`

	req := httptest.NewRequest("POST", "/transfer", strings.NewReader(requestBodyJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-jwt-token", createTestJWT(t, 9901, "user"))

	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/transfer", jwtAuthMiddleware(makeHttpHandleFunc(server.handleTransfer))).Methods("POST")

	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var returnedAcc *Account
	err := json.Unmarshal(recorder.Body.Bytes(), &returnedAcc)
	require.NoError(t, err)
	assert.Equal(t, fromAccount.Number, returnedAcc.Number)
	assert.Equal(t, fromAccount.Balance, returnedAcc.Balance)
	assert.Equal(t, fromAccount.FirstName, returnedAcc.FirstName)
	assert.Equal(t, fromAccount.LastName, returnedAcc.LastName)
}

func createTestJWT(t *testing.T, accountNumber int64, role string) string {
	secret := os.Getenv("JWT_SECRET")

	claims := &jwt.MapClaims{
		"role":          role,
		"expiresAt":     jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		"accountNumber": float64(accountNumber),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	return tokenString
}
