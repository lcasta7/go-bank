package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sethvargo/go-retry"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	TransferMoney(*Account, *Account, uint64) error
	GetAccountByNumber(int64) (*Account, error)
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
                  encrypted_password VARCHAR(100),
                  balance BIGINT,
                  created_at timestamp DEFAULT NOW()
           )`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgressStore) CreateAccount(acc *Account) error {
	query := `INSERT INTO account
(first_name, last_name, number, encrypted_password, balance, created_at)
VALUES
($1, $2, $3, $4, $5, $6)
RETURNING ID`

	err := s.db.QueryRow(query, acc.FirstName,
		acc.LastName, acc.Number, acc.EncryptedPassword,
		acc.Balance, acc.CreatedAt).Scan(&acc.Id)

	return err
}

func (s *PostgressStore) DeleteAccount(id int) error {
	result, err := s.db.Exec("DELETE FROM ACCOUNT WHERE ID = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Could not delete account with id=%d", id)
	}

	return err
}

func (s *PostgressStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgressStore) TransferMoney(fromAcc *Account, toAcc *Account, amount uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b := retry.NewFibonacci(10 * time.Millisecond)
	b = retry.WithMaxDuration(5*time.Second, b)

	err := retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Attempt to execute operations within transaction
		// Use defer with a function to handle rollback logic
		committed := false
		defer func() {
			if !committed {
				tx.Rollback()
			}
		}()

		// Update first account
		firstAccBalanceUpdate := fromAcc.Balance - amount
		_, err = tx.ExecContext(ctx,
			"UPDATE ACCOUNT SET balance = $1 WHERE number = $2",
			firstAccBalanceUpdate, fromAcc.Number)
		if err != nil {
			return fmt.Errorf("failed to update source account: %w", err)
		}

		secondAccBalanceUpdate := toAcc.Balance + amount
		_, err = tx.ExecContext(ctx,
			"UPDATE ACCOUNT SET balance = $1 WHERE number = $2",
			secondAccBalanceUpdate, toAcc.Number)
		if err != nil {
			return fmt.Errorf("failed to update destination account: %w", err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		committed = true
		return nil
	})
	if err != nil {
		return fmt.Errorf("transfer failed after retries: %w", err)
	}
	return nil
}

func (s *PostgressStore) GetAccountByNumber(number int64) (*Account, error) {
	rows, err := s.db.Query("SELECT * FROM ACCOUNT WHERE NUMBER = $1", number)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account number not found for number %d", number)
}

func (s *PostgressStore) GetAccountById(id int) (*Account, error) {
	rows, err := s.db.Query("SELECT * FROM ACCOUNT WHERE ID = $1", id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account with id %d not found", id)
}

func (s *PostgressStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT * FROM ACCOUNT")

	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, error := scanIntoAccount(rows)

		if error != nil {
			return nil, error
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.Id,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.EncryptedPassword,
		&account.Balance,
		&account.CreatedAt)

	return account, err
}
