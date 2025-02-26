package main

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountById(int) (*Account, error)
	GetAccounts() ([]*Account, error)
}

// what other db could I use?
type PostgressStore struct {
	db *sql.DB
}

func NewProstgressStore() (*PostgressStore, error) {
	connStr := "user=postgres dbname=postgres password=gobank sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if pingErr := db.Ping(); pingErr != nil {
		return nil, pingErr
	}

	return &PostgressStore{
		db: db,
	}, nil
}

func (s *PostgressStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgressStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS account (
                  id SERIAL PRIMARY KEY,
                  first_name VARCHAR(50),
                  last_name VARCHAR(50),
                  number BIGINT,
                  balance BIGINT,
                  created_at timestamp DEFAULT NOW()
           )`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgressStore) CreateAccount(acc *Account) error {
	query := `INSERT INTO account
(first_name, last_name, number, balance, created_at)
VALUES
($1, $2, $3, $4, $5)
RETURNING ID`

	err := s.db.QueryRow(query, acc.FirstName,
		acc.LastName, acc.Number,
		acc.Balance, acc.CreatedAt).Scan(&acc.Id)

	return err
}

func (s *PostgressStore) DeleteAccount(id int) error {
	return nil
}
func (s *PostgressStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgressStore) GetAccountById(id int) (*Account, error) {
	return nil, nil
}

func (s *PostgressStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT * FROM ACCOUNT")

	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account := new(Account)
		err := rows.Scan(
			&account.Id,
			&account.FirstName,
			&account.LastName,
			&account.Number,
			&account.Balance,
			&account.CreatedAt)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}
