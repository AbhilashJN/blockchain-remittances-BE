package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AbhilashJN/blockchain-remittances-BE/db"
)

func main() {

	for _, dbname := range []string{"SBI.db", "JPM.db", "CustomerPool.db"} {
		err := os.Remove(dbname)
		if err == nil {
			continue
		}
		err = err.(*os.PathError)
		fmt.Println(err)
	}

	for _, bankName := range []string{"SBI", "JPM"} {
		if err := db.CreateDBForBank(bankName); err != nil {
			log.Fatal(err)
		}
	}

	if err := db.CreateCustomerPoolDB(); err != nil {
		log.Fatal(err)
	}
}
