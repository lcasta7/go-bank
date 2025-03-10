package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"time"
)

type ApiServer struct {
	listenAddr string
	store      Storage
}

func NewApiServer(listenAddr string, store Storage) *ApiServer {
	return &ApiServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *ApiServer) Run() {
	router := mux.NewRouter()

	//this should be admin endpoint
	router.HandleFunc("/account", makeHttpHandleFunc(s.handleGetAccounts)).Methods("GET")    //get all accounts
	router.HandleFunc("/account", makeHttpHandleFunc(s.handleCreateAccount)).Methods("POST") //create new account

	//user api
	router.HandleFunc("/login", makeHttpHandleFunc(s.handleLogin)).Methods("POST")
	router.HandleFunc("/account/{id}", withJwtAuth(makeHttpHandleFunc(s.handleGetAccountById), s.store)).Methods("GET")
	router.HandleFunc("/account/{id}", makeHttpHandleFunc(s.handleDeleteAccount)).Methods("DELETE")
	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer)).Methods("POST")

	log.Println("Starting the server port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *ApiServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		fmt.Println("Unable to parse request")
		return fmt.Errorf("Not Authenticated")
	}

	//handle acc
	acc, err := s.store.GetAccountByNumber(req.Number)

	if err != nil {
		fmt.Println("Error retrieving account")
		return fmt.Errorf("Not Authenticated")
	}

	//verify that the passwords match
	if err := acc.ValidatePassword(req.Password); err != nil {
		fmt.Println("Error validating password")
		return fmt.Errorf("Not Authenticated")
	}

	token, err := createJwt(acc)
	if err != nil {
		fmt.Println("Error creating JWT")
		return fmt.Errorf("Not Authenticated")
	}

	resp := LoginResponse{
		Number: acc.Number,
		Token:  token,
	}

	return WriteJson(w, http.StatusOK, resp)
}

func (s *ApiServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		fmt.Println("Error retrieving accounts")
		return err
	}

	return WriteJson(w, http.StatusOK, accounts)
}

func (s *ApiServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	parameter, err := getParameter(r, "id")
	if err != nil {
		fmt.Println("Error retrieving parameter")
		return err
	}

	id, err := strconv.Atoi(parameter)
	if err != nil {
		fmt.Printf("Unable to convert parameter %s", parameter)
		return err
	}

	account, err := s.store.GetAccountById(id)
	if err != nil {
		fmt.Printf("Error retrieving account for %d", id)
		return err
	}

	return WriteJson(w, http.StatusOK, account)
}

func (s *ApiServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	accRequest := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(accRequest); err != nil {
		return err
	}
	defer r.Body.Close()

	account, err := NewAccount(accRequest.FirstName, accRequest.LastName, accRequest.Password, accRequest.Balance)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(account); err != nil {
		fmt.Println("Error creating account")
		return err
	}

	fmt.Printf("Welcome %s, with balance=%d", accRequest.FirstName, accRequest.Balance)

	return WriteJson(w, http.StatusOK, account)

}

func (s *ApiServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	parameter, err := getParameter(r, "id")
	if err != nil {
		fmt.Println("Error retrieving parameter")
		return err
	}

	id, err := strconv.Atoi(parameter)
	if err != nil {
		fmt.Printf("Unable to convert parameter %s", parameter)
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return fmt.Errorf("Could not delete account with id=%d", id)
	}

	return WriteJson(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *ApiServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferRequest := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferRequest); err != nil {
		return err
	}
	defer r.Body.Close()

	return WriteJson(w, http.StatusOK, transferRequest)
}

// least important functions should go to the bottom
func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// interesting, notice the decoration of this.
// now it can be modified before calling the actual function in the router
func withJwtAuth(handlerFunc http.HandlerFunc, storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJwt(tokenString)
		if err != nil {
			fmt.Println("Unable to validate token")
			WriteJson(w, http.StatusForbidden, ApiError{Error: "permission denied"})
			return
		}

		if !token.Valid {
			fmt.Println("Invalid Token")
			WriteJson(w, http.StatusForbidden, ApiError{Error: "permission denied"})
			return
		}

		paramId, err := getParameter(r, "id")
		if err != nil {
			fmt.Println("No parameter Id")
			WriteJson(w, http.StatusForbidden, ApiError{Error: "permission denied"})
			return
		}
		accountId, err := strconv.Atoi(paramId)
		if err != nil {
			fmt.Println("Invalid Id")
			WriteJson(w, http.StatusForbidden, ApiError{Error: "permission denied"})
			return
		}

		account, err := storage.GetAccountById(accountId)
		if err != nil {
			fmt.Println("Could not retrieve for id")
			WriteJson(w, http.StatusForbidden, ApiError{Error: "permission denied"})
			return
		}

		//get user claims
		claims := token.Claims.(jwt.MapClaims)
		claimAccountNumber, ok := claims["accountNumber"].(float64)
		if !ok {
			fmt.Println("Invalid claim type for accountNumber")
			WriteJson(w, http.StatusForbidden, ApiError{Error: "invalid token"})
			return
		}

		if account.Number != int64(claimAccountNumber) {
			fmt.Println("Insufficient Permissions")
			WriteJson(w, http.StatusForbidden, ApiError{Error: "permission denied"})
			return
		}

		handlerFunc(w, r)
	}
}

func validateJwt(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func createJwt(account *Account) (string, error) {
	// Create the Claims
	claims := &jwt.MapClaims{
		"expiresAt":     jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		"accountNumber": account.Number,
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET environment variable not set")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

// my functions are of this type by virtue of the signature
type apiFunc func(w http.ResponseWriter, r *http.Request) error
type ApiError struct {
	Error string `json:"error"`
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJson(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}

	}
}

func getParameter(r *http.Request, field string) (string, error) {
	parameter, ok := mux.Vars(r)[field]

	if !ok {
		return "", fmt.Errorf("invalid parameter: %s", field)
	}

	return parameter, nil
}
