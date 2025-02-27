package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount)).Methods("GET")
	router.HandleFunc("/account/{id}", makeHttpHandleFunc(s.handleGetAccountById)).Methods("GET")
	router.HandleFunc("/account/{id}", makeHttpHandleFunc(s.handleDeleteAccount)).Methods("DELETE")
	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer)).Methods("POST")

	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", makeHttpHandleFunc(s.handleAccount))

	log.Println("Starting the server port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *ApiServer) handleAccount(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("Not allowed %s", r.Method)
}

func (s *ApiServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
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

	account := NewAccount(accRequest.FirstName, accRequest.LastName, accRequest.Balance)

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
