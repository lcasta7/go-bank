package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"time"

	jwt "github.com/golang-jwt/jwt/v5"
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

	//admin endpoints
	router.HandleFunc("/accounts",
		jwtAuthMiddleware(makeHttpHandleFunc(s.handleGetAccounts))).Methods("GET")
	router.HandleFunc("/account",
		jwtAuthMiddleware(makeHttpHandleFunc(s.handleCreateAccount))).Methods("POST")
	router.HandleFunc("/account/{id}",
		jwtAuthMiddleware(makeHttpHandleFunc(s.handleDeleteAccount))).Methods("DELETE")

	//user endpoints
	router.HandleFunc("/login", makeHttpHandleFunc(s.handleLogin)).Methods("POST")
	router.HandleFunc("/account",
		jwtAuthMiddleware(makeHttpHandleFunc(s.handleGetAccountByNumber))).Methods("GET")
	router.HandleFunc("/transfer",
		jwtAuthMiddleware(makeHttpHandleFunc(s.handleTransfer))).Methods("POST")

	//start server
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
	_, err := decodeAndValidateRequest[GetAccountRequest](r, "admin")

	if err != nil {
		fmt.Println("Error validating request")
		return fmt.Errorf("Error processing request")
	}

	accounts, err := s.store.GetAccounts()
	if err != nil {
		fmt.Println("Error retrieving accounts")
		return err
	}

	return WriteJson(w, http.StatusOK, accounts)
}

func decodeAndValidateRequest[T any](r *http.Request, requestType string) (*T, error) {
	// Create a new instance of the request type
	req := new(T)
	authorizedAccountNumber := r.Context().Value("authorizedAccountNumber").(int64)
	role, ok := r.Context().Value("role").(string)

	// handle admin request
	if requestType == "admin" {
		if !ok || role != "admin" {
			return nil, fmt.Errorf("Insufficient permissions: admin role required")
		}
	}

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Validate the account number
	reqWithAccount, ok := any(req).(HttpRequest)
	if !ok {
		return nil, fmt.Errorf("request type does not implement RequestWithAccount interface")
	}

	if reqWithAccount.GetAccountNumber() != authorizedAccountNumber {
		return nil, fmt.Errorf("access denied: account numbers do not match")
	}

	return req, nil
}

func (s *ApiServer) handleGetAccountByNumber(w http.ResponseWriter, r *http.Request) error {
	getAccountRequest, err := decodeAndValidateRequest[GetAccountRequest](r, "user")

	if err != nil {
		fmt.Println("Error validating request")
		return fmt.Errorf("Error processing request")
	}

	account, err := s.store.GetAccountByNumber(getAccountRequest.Number)
	if err != nil {
		fmt.Println("Error retrieving account from db")
		return fmt.Errorf("Error processing request")
	}

	return WriteJson(w, http.StatusOK, account)
}

func (s *ApiServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	//todo check that it's admin
	accRequest, err := decodeAndValidateRequest[CreateAccountRequest](r, "admin")
	if err != nil {
		return err
	}

	//default the role to user
	if accRequest.Role == "" {
		accRequest.Role = "user"
	}

	account, err := NewAccount(accRequest.FirstName, accRequest.LastName, accRequest.Password, accRequest.Role, accRequest.Balance)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(account); err != nil {
		fmt.Println("Error creating account")
		return err
	}

	return WriteJson(w, http.StatusOK, account)
}

func (s *ApiServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	_, err := decodeAndValidateRequest[DeleteAccountRequest](r, "admin")

	if err != nil {
		fmt.Println("Error decoding delete request")
		return WriteJson(w, http.StatusBadRequest, fmt.Errorf("Unable to delete account"))
	}

	parameter, err := getParameter(r, "id")
	if err != nil {
		fmt.Println("Error retrieving parameter")
		return WriteJson(w, http.StatusBadRequest, fmt.Errorf("Unable to delete account"))
	}

	id, err := strconv.Atoi(parameter)
	if err != nil {
		fmt.Printf("Unable to convert parameter %s", parameter)
		return WriteJson(w, http.StatusBadRequest, fmt.Errorf("Unable to delete account"))
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return WriteJson(w, http.StatusBadRequest, fmt.Errorf("Unable to delete account"))
	}

	return WriteJson(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *ApiServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	getTransferRequest, err := decodeAndValidateRequest[TransferRequest](r, "user")

	if err != nil {
		fmt.Println("Error decoding into transfer request")
		return WriteJson(w, http.StatusBadRequest, ApiError{Error: "could not complete request"})
	}

	fromAccount, err := s.store.GetAccountByNumber(getTransferRequest.FromNumber)
	if err != nil {
		fmt.Println("Error retrieving source account")
		return WriteJson(w, http.StatusInternalServerError, ApiError{Error: "could not complete request"})
	}
	if fromAccount.Balance < getTransferRequest.Amount {
		fmt.Println("Insufficient funds in source account")
		return WriteJson(w, http.StatusBadRequest, ApiError{Error: "could not complete request"})
	}

	toAccount, err := s.store.GetAccountByNumber(getTransferRequest.ToNumber)
	if err != nil {
		fmt.Println("Could not retrieve destination account")
		return WriteJson(w, http.StatusInternalServerError, ApiError{Error: "could not complete request"})
	}
	if err := s.store.TransferMoney(fromAccount, toAccount, getTransferRequest.Amount); err != nil {
		fmt.Println("Could complete the transfer")
		return WriteJson(w, http.StatusInternalServerError, ApiError{Error: "could not complete request"})
	}

	fromAccountUpdated, err := s.store.GetAccountByNumber(getTransferRequest.FromNumber)
	if err != nil {
		fmt.Println("Could not retrieve updated from account")
		return WriteJson(w, http.StatusInternalServerError, ApiError{Error: "Transfer successful, but could not retrieve updated account"})
	}

	return WriteJson(w, http.StatusOK, fromAccountUpdated)
}

// least important functions should go to the bottom
func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func getClaimsMap(w http.ResponseWriter, r *http.Request) (jwt.MapClaims, error) {
	tokenString := r.Header.Get("x-jwt-token")
	if tokenString == "" {
		WriteJson(w, http.StatusUnauthorized, ApiError{Error: "authentication required"})
		return nil, fmt.Errorf("Not logged in.")
	}

	token, err := validateJwt(tokenString)
	if err != nil || !token.Valid {
		WriteJson(w, http.StatusForbidden, ApiError{Error: "permission denied"})
		return nil, fmt.Errorf("Can't validate token.")
	}

	//check the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		WriteJson(w, http.StatusUnauthorized, ApiError{Error: "permission denied"})
		return nil, fmt.Errorf("Invalid claims format")
	}

	return claims, nil
}

// this will only check for user claims
func jwtAuthMiddleware(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := getClaimsMap(w, r)
		if err != nil {
			fmt.Println("Invalid claims format")
			WriteJson(w, http.StatusUnauthorized, ApiError{Error: "permission denied"})
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			fmt.Println("Role not specified, defaulting to user")
			role = "user"
		}

		claimAccountNumber, ok := claims["accountNumber"].(float64)
		if !ok {
			fmt.Println("Invalid claim type for accountNumber")
			WriteJson(w, http.StatusForbidden, ApiError{Error: "permission denied"})
			return
		}

		ctx := context.WithValue(r.Context(), "role", string(role))
		ctx = context.WithValue(ctx, "authorizedAccountNumber", int64(claimAccountNumber))

		handlerFunc(w, r.WithContext(ctx))
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
		"role":          account.Role,
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
