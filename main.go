package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

// func lumenTransaction() {
// 	sender := account.GetCreatedAccounts().PersonA
// 	receiver := account.GetCreatedAccounts().PersonB
// 	var messages = make(chan string, 4)
// 	go transaction.WatchLiveActivityOf(sender.Address, messages)
// 	go transaction.WatchLiveActivityOf(receiver.Address, messages)
// 	transaction.Transact(sender, receiver, "555")
// 	println(<-messages)
// 	println(<-messages)
// 	println("Sender to receiver done")
// 	// transaction.Transact(receiver, sender, "555")
// 	// go transaction.WatchLiveActivityOf(sender.Address, messages)
// 	// go transaction.WatchLiveActivityOf(receiver.Address, messages)
// 	// println(<-messages)
// 	// println(<-messages)
// 	// println("receiver to sender done")
// 	println("all routines exited")
// }

// func customAssetTransaction() {
// 	personA := account.GetCreatedAccounts().PersonA
// 	personB := account.GetCreatedAccounts().PersonB
// 	// fmt.Printf("issuer before keyparse: %+v\n", personA)
// 	// fmt.Printf("recipient before keyparse: %+v\n", personB)
// 	issuerSeed := personA.Seed
// 	recipientSeed := personB.Seed

// 	// Keys for accounts to issue and receive the new asset
// 	issuer, err := keypair.Parse(issuerSeed)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	recipient, err := keypair.Parse(recipientSeed)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// fmt.Printf("issuer after keyparse: %+v\n", issuer.Address())
// 	// fmt.Printf("recipient after keyparse: %+v\n", recipient.Address())

// 	var messages = make(chan string, 2)
// 	go transaction.WatchLiveActivityOf(issuer.Address(), messages)
// 	go transaction.WatchLiveActivityOf(recipient.Address(), messages)

// 	// Create an object to represent the new asset
// 	USDT := build.CreditAsset("USDT", issuer.Address())

// 	// First, the receiving account must trust the asset
// 	trustTx, err := build.Transaction(
// 		build.SourceAccount{AddressOrSeed: recipient.Address()},
// 		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
// 		build.TestNetwork,
// 		build.Trust(USDT.Code, USDT.Issuer),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	trustTxe, err := trustTx.Sign(recipientSeed)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	trustTxeB64, err := trustTxe.Base64()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	_, err = horizon.DefaultTestNetClient.SubmitTransaction(trustTxeB64)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Second, the issuing account actually sends a payment using the asset
// 	paymentTx, err := build.Transaction(
// 		build.SourceAccount{AddressOrSeed: issuer.Address()},
// 		build.TestNetwork,
// 		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
// 		build.Payment(
// 			build.Destination{AddressOrSeed: recipient.Address()},
// 			build.CreditAmount{Code: "USDT", Issuer: issuer.Address(), Amount: "10"},
// 		),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	paymentTxe, err := paymentTx.Sign(issuerSeed)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	paymentTxeB64, err := paymentTxe.Base64()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	_, err = horizon.DefaultTestNetClient.SubmitTransaction(paymentTxeB64)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	println(<-messages)
// 	println(<-messages)
// 	println("all routines exited")
// }

// func trustAsset(recipientKP keypair.KP, recipientSeed string, asset build.Asset) {
// 	trustTx, err := build.Transaction(
// 		build.SourceAccount{AddressOrSeed: recipientKP.Address()},
// 		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
// 		build.TestNetwork,
// 		build.Trust(asset.Code, asset.Issuer),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	trustTxe, err := trustTx.Sign(recipientSeed)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	trustTxeB64, err := trustTxe.Base64()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	_, err = horizon.DefaultTestNetClient.SubmitTransaction(trustTxeB64)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func getIssuer() (keypair.KP, string) {
// 	personA := account.GetCreatedAccounts().PersonA
// 	issuerSeed := personA.Seed
// 	issuerKP, err := keypair.Parse(issuerSeed)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return issuerKP, issuerSeed
// }

// func getRecipient() (keypair.KP, string) {
// 	personB := account.GetCreatedAccounts().PersonB
// 	recipientSeed := personB.Seed
// 	recipientKP, err := keypair.Parse(recipientSeed)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return recipientKP, recipientSeed
// }

// func createAndSendCustomTokenFromAtoB(issuerKP, recipientKP keypair.KP, issuerSeed, recipientSeed string, customAsset build.Asset, assetAmout string) {

// 	fmt.Println("_____________________________________before_____________________________________")
// 	account.PrintAccountDetails(issuerKP.Address())
// 	account.PrintAccountDetails(recipientKP.Address())
// 	fmt.Println("_____________________________________before_____________________________________")

// 	trustAsset(recipientKP, recipientSeed, customAsset)

// 	// sendAssetFromAtoB(issuerKP, recipientKP, issuerSeed, customAsset, assetAmout)

// 	fmt.Println("_____________________________________after_______________________________________")
// 	account.PrintAccountDetails(issuerKP.Address())
// 	account.PrintAccountDetails(recipientKP.Address())
// 	fmt.Println("_____________________________________after_______________________________________")
// }

// // PrintBalencesOfSIDaccounts returns
// func PrintBalencesOfSIDaccounts(stellarAdresses *data.StellarAddresses) {
// 	fmt.Println("SOURCE ACCOUNT START------------------------------------------------\n-")
// 	account.PrintAccountDetails(stellarAdresses.Source)
// 	fmt.Println("SOURCE ACCOUNT END------------------------------------------------\n-")
// 	fmt.Println("ISSUER ACCOUNT START:------------------------------------------------\n-")
// 	account.PrintAccountDetails(stellarAdresses.Issuer)
// 	fmt.Println("ISSUER ACCOUNT END------------------------------------------------\n-")
// 	fmt.Println("DISTRIBUTION ACCOUNT START:------------------------------------------------\n-")
// 	account.PrintAccountDetails(stellarAdresses.Distributor)
// 	fmt.Println("DISTRIBUTION ACCOUNT END------------------------------------------------\n-")
// }

var filePath = flag.String("configFile", "./ALPHAconfig.yml", "config filepath")

// Keys is
type Keys struct {
	Source, Issuer, Distributor string
}

// DBconnectionParams is
type DBconnectionParams struct {
	Host, Port, User, Password, DbName string
}

// BankConfig is
type BankConfig struct {
	Name               string
	Port               string
	StellarSeeds       Keys
	StellarAddresses   Keys
	StellarAssetCode   string
	DBconnectionParams DBconnectionParams
	DB                 *gorm.DB
	BankPoolAccID      string
}

func readConfig(configFilePath string) BankConfig {
	viper.SetConfigFile(configFilePath)

	// Searches for config file in given paths and read it
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Confirm which config file is used
	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())

	var bankConfig BankConfig

	if err := viper.Unmarshal(&bankConfig); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	return bankConfig
}

func init() {

}

func main() {
	// err := onboarding.OnboardBank("ALPHAT", "500000000", "100", "100")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	flag.Parse()
	bankConfig := readConfig(*filePath)
	spew.Dump(bankConfig)

	dbConn := bankConfig.DBconnectionParams
	dbConnParam := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", dbConn.Host, dbConn.Port, dbConn.User, dbConn.DbName, dbConn.Password)

	db, err := gorm.Open("postgres", dbConnParam)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	bankConfig.DB = db

	go listenForPayments(bankConfig)
	StartServer(bankConfig)
}
