package main

import (
	"flag"
	"fmt"
	"log"
)

func seedAccount(store Storage, fname, lname, pw, role string) *Account {

	acc, err := NewAccount(fname, lname, pw, role, 0)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)

	}

	fmt.Println("new account -> ", acc.Number)

	return acc

}

// seeding a test db
func seedAccounts(s Storage) {
	seedAccount(s, "luis", "cast", "password", "user")
}

func createAdminAccount(s Storage, firstName, lastName, password string) {
	seedAccount(s, firstName, lastName, password, "admin")
}

func main() {
	seed := flag.Bool("seed", false, "seed the db")
	createAdmin := flag.Bool("create-admin", false, "create an admin account")
	firstName := flag.String("first_name", "", "first name for admin account")
	lastName := flag.String("last_name", "", "last name for admin account")
	password := flag.String("password", "", "password for admin account")
	flag.Parse()

	store, err := NewProstgressStore()

	if err != nil {
		log.Fatal("Error initializing the db")
	}

	if err := store.Init(); err != nil {
		log.Fatal("Error creating table ", err)
	}

	if *seed {
		fmt.Println("Seeding the database")
		seedAccounts(store)
	}

	if *createAdmin {
		if *firstName == "" || *lastName == "" || *password == "" {
			log.Fatal("insufficient fields")
		}

		createAdminAccount(store, *firstName, *lastName, *password)
		log.Println("Admin account created successfully")
		return
	}

	server := NewApiServer(":3000", store)
	server.Run()
}
