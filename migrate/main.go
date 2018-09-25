package main

import (
	"log"

	"github.com/AbhilashJN/blockchain-remittances-BE/db"
)

func main() {

	err := db.CreateDBForBank("SBI")
	if err != nil {
		log.Fatal(err)
	}

	err = db.CreateDBForBank("JPM")
	if err != nil {
		log.Fatal(err)
	}

	db.C

}
