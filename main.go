package main

import (
	"flag"
	"fmt"
	"log"
)

func seedAccount(store Storage, fname string, lname string, pw string) *Account {

	acc, err := NewAccount(fname, lname, pw, 100)
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
	seedAccount(s, "luis", "cast", "password")
}

func main() {
	seed := flag.Bool("seed", false, "seed the db")
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

	server := NewApiServer(":3000", store)
	server.Run()
}
