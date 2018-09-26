package main

import (
	"fmt"
	"log"

	"github.com/AbhilashJN/blockchain-remittances-BE/account"
	"github.com/AbhilashJN/blockchain-remittances-BE/bank"
	"github.com/AbhilashJN/blockchain-remittances-BE/db"
	"github.com/AbhilashJN/blockchain-remittances-BE/transaction"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
)

func lumenTransaction() {
	sender := account.GetCreatedAccounts().PersonA
	receiver := account.GetCreatedAccounts().PersonB
	var messages = make(chan string, 4)
	go transaction.WatchLiveActivityOf(sender.Address, messages)
	go transaction.WatchLiveActivityOf(receiver.Address, messages)
	transaction.Transact(sender, receiver, "555")
	println(<-messages)
	println(<-messages)
	println("Sender to receiver done")
	// transaction.Transact(receiver, sender, "555")
	// go transaction.WatchLiveActivityOf(sender.Address, messages)
	// go transaction.WatchLiveActivityOf(receiver.Address, messages)
	// println(<-messages)
	// println(<-messages)
	// println("receiver to sender done")
	println("all routines exited")
}

func customAssetTransaction() {
	personA := account.GetCreatedAccounts().PersonA
	personB := account.GetCreatedAccounts().PersonB
	// fmt.Printf("issuer before keyparse: %+v\n", personA)
	// fmt.Printf("recipient before keyparse: %+v\n", personB)
	issuerSeed := personA.Seed
	recipientSeed := personB.Seed

	// Keys for accounts to issue and receive the new asset
	issuer, err := keypair.Parse(issuerSeed)
	if err != nil {
		log.Fatal(err)
	}
	recipient, err := keypair.Parse(recipientSeed)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("issuer after keyparse: %+v\n", issuer.Address())
	// fmt.Printf("recipient after keyparse: %+v\n", recipient.Address())

	var messages = make(chan string, 2)
	go transaction.WatchLiveActivityOf(issuer.Address(), messages)
	go transaction.WatchLiveActivityOf(recipient.Address(), messages)

	// Create an object to represent the new asset
	USDT := build.CreditAsset("USDT", issuer.Address())

	// First, the receiving account must trust the asset
	trustTx, err := build.Transaction(
		build.SourceAccount{AddressOrSeed: recipient.Address()},
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.TestNetwork,
		build.Trust(USDT.Code, USDT.Issuer),
	)
	if err != nil {
		log.Fatal(err)
	}
	trustTxe, err := trustTx.Sign(recipientSeed)
	if err != nil {
		log.Fatal(err)
	}
	trustTxeB64, err := trustTxe.Base64()
	if err != nil {
		log.Fatal(err)
	}
	_, err = horizon.DefaultTestNetClient.SubmitTransaction(trustTxeB64)
	if err != nil {
		log.Fatal(err)
	}

	// Second, the issuing account actually sends a payment using the asset
	paymentTx, err := build.Transaction(
		build.SourceAccount{AddressOrSeed: issuer.Address()},
		build.TestNetwork,
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.Payment(
			build.Destination{AddressOrSeed: recipient.Address()},
			build.CreditAmount{Code: "USDT", Issuer: issuer.Address(), Amount: "10"},
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	paymentTxe, err := paymentTx.Sign(issuerSeed)
	if err != nil {
		log.Fatal(err)
	}
	paymentTxeB64, err := paymentTxe.Base64()
	if err != nil {
		log.Fatal(err)
	}
	_, err = horizon.DefaultTestNetClient.SubmitTransaction(paymentTxeB64)
	if err != nil {
		log.Fatal(err)
	}

	println(<-messages)
	println(<-messages)
	println("all routines exited")
}

func trustAsset(recipientKP keypair.KP, recipientSeed string, asset build.Asset) {
	trustTx, err := build.Transaction(
		build.SourceAccount{AddressOrSeed: recipientKP.Address()},
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.TestNetwork,
		build.Trust(asset.Code, asset.Issuer),
	)
	if err != nil {
		log.Fatal(err)
	}
	trustTxe, err := trustTx.Sign(recipientSeed)
	if err != nil {
		log.Fatal(err)
	}
	trustTxeB64, err := trustTxe.Base64()
	if err != nil {
		log.Fatal(err)
	}
	_, err = horizon.DefaultTestNetClient.SubmitTransaction(trustTxeB64)
	if err != nil {
		log.Fatal(err)
	}
}

func buildAsset(issuerKP keypair.KP, assetCode string) build.Asset {
	asset := build.CreditAsset(assetCode, issuerKP.Address())
	return asset
}

func getIssuer() (keypair.KP, string) {
	personA := account.GetCreatedAccounts().PersonA
	issuerSeed := personA.Seed
	issuerKP, err := keypair.Parse(issuerSeed)
	if err != nil {
		log.Fatal(err)
	}
	return issuerKP, issuerSeed
}

func getKeyPair(seed string) (keypair.KP, error) {
	accKeyPair, err := keypair.Parse(seed)
	if err != nil {
		return nil, err
	}
	return accKeyPair, nil
}

func getRecipient() (keypair.KP, string) {
	personB := account.GetCreatedAccounts().PersonB
	recipientSeed := personB.Seed
	recipientKP, err := keypair.Parse(recipientSeed)
	if err != nil {
		log.Fatal(err)
	}
	return recipientKP, recipientSeed
}

func sendAssetFromAtoB(A, B keypair.KP, Aseed string, asset build.Asset, amount string) {
	paymentTx, err := build.Transaction(
		build.SourceAccount{AddressOrSeed: A.Address()},
		build.TestNetwork,
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.Payment(
			build.Destination{AddressOrSeed: B.Address()},
			build.CreditAmount{Code: asset.Code, Issuer: asset.Issuer, Amount: amount},
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	paymentTxe, err := paymentTx.Sign(Aseed)
	if err != nil {
		log.Fatal(err)
	}
	paymentTxeB64, err := paymentTxe.Base64()
	if err != nil {
		log.Fatal(err)
	}
	_, err = horizon.DefaultTestNetClient.SubmitTransaction(paymentTxeB64)
	if err != nil {
		log.Fatal(err)
	}

}

func createAndSendCustomTokenFromAtoB(issuerKP, recipientKP keypair.KP, issuerSeed, recipientSeed string, customAsset build.Asset, assetAmout string) {

	fmt.Println("_____________________________________before shit_____________________________________")
	account.PrintAccountDetails(issuerKP.Address())
	account.PrintAccountDetails(recipientKP.Address())
	fmt.Println("_____________________________________before shit_____________________________________")

	trustAsset(recipientKP, recipientSeed, customAsset)

	sendAssetFromAtoB(issuerKP, recipientKP, issuerSeed, customAsset, assetAmout)

	fmt.Println("_____________________________________after shit_______________________________________")
	account.PrintAccountDetails(issuerKP.Address())
	account.PrintAccountDetails(recipientKP.Address())
	fmt.Println("_____________________________________after shit_______________________________________")
}

// GetSIDkeyPairsOfBank returns
func GetSIDkeyPairsOfBank(stellarAddressesOfBank *db.StellarAddressesOfBank) bank.SIDKeyPairs {
	SourceKeyPair, err := getKeyPair(stellarAddressesOfBank.SourceSeed)
	if err != nil {
		log.Fatal(err)
	}
	IssuerKeyPair, err := getKeyPair(stellarAddressesOfBank.IssuerSeed)
	if err != nil {
		log.Fatal(err)
	}
	DistributorKeyPair, err := getKeyPair(stellarAddressesOfBank.DistributorSeed)
	if err != nil {
		log.Fatal(err)
	}
	return bank.SIDKeyPairs{Source: SourceKeyPair, Issuer: IssuerKeyPair, Distributor: DistributorKeyPair}
}

// PrintBalencesOfSIDaccounts returns
func PrintBalencesOfSIDaccounts(stellarAdressesOfBank bank.SIDKeyPairs) {
	fmt.Println("SOURCE ACCOUNT START------------------------------------------------\n-")
	account.PrintAccountDetails(stellarAdressesOfBank.Source.Address())
	fmt.Println("SOURCE ACCOUNT END------------------------------------------------\n-")
	fmt.Println("ISSUER ACCOUNT START:------------------------------------------------\n-")
	account.PrintAccountDetails(stellarAdressesOfBank.Issuer.Address())
	fmt.Println("ISSUER ACCOUNT END------------------------------------------------\n-")
	fmt.Println("DISTRIBUTION ACCOUNT START:------------------------------------------------\n-")
	account.PrintAccountDetails(stellarAdressesOfBank.Distributor.Address())
	fmt.Println("DISTRIBUTION ACCOUNT END------------------------------------------------\n-")
}

func main() {
	// lumenTransaction()
	// customAssetTransaction()
	// issuerKP, issuerSeed := getIssuer()
	// recipientKP, recipientSeed := getRecipient()
	// customAsset := buildAsset(issuerKP, "ABC")
	// createAndSendCustomTokenFromAtoB(issuerKP, recipientKP, issuerSeed, recipientSeed, customAsset, "999")
	// sendAssetFromAtoB(recipientKP, issuerKP, recipientSeed, customAsset, "999")
	// account.PrintAccountDetails(issuerKP.Address())
	// account.PrintAccountDetails(recipientKP.Address())

	// err := bank.OnboardBank("JPMORGAN", "JPMRT")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// stellarSeedsOfJPM, err := db.RetreiveStellarAddressesOfBank("JPMORGAN")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%+v \n", *stellarSeedsOfJPM)

	stellarSeedsOfSBI, err := db.ReadStellarAddressesOfBank("SBI")
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("%+v \n", stellarSeedsOfSBI)

	stellarSeedsOfJPM, err := db.ReadStellarAddressesOfBank("JPM")
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("%+v \n", stellarSeedsOfJPM)
	// stellarAddressesOfSBI := GetSIDkeyPairsOfBank(stellarSeedsOfSBI)
	// stellarAddressesOfJPM := GetSIDkeyPairsOfBank(stellarSeedsOfJPM)

	// PrintBalencesOfSIDaccounts(stellarAddressesOfSBI)
	// PrintBalencesOfSIDaccounts(stellarAddressesOfJPM)
	// JpmtAsset := buildAsset(stellarAddressesOfJPM.Issuer, "JPMT")
	// SbitAsset := buildAsset(stellarAddressesOfSBI.Issuer, "SBIT")

	// if err := bank.CreateTrust(stellarSeedsOfSBI.DistributorSeed, stellarSeedsOfJPM.IssuerSeed, "JPMT"); err != nil {
	// 	log.Fatal(err)
	// }

	// if err := bank.CreateTrust(stellarSeedsOfJPM.DistributorSeed, stellarSeedsOfSBI.IssuerSeed, "SBIT"); err != nil {
	// 	log.Fatal(err)
	// }

	// bank.IssueToDistribAccount(stellarSeedsOfJPM.DistributorSeed, stellarSeedsOfJPM.IssuerSeed, "JPMT", "100")
	// bank.IssueToDistribAccount(stellarSeedsOfSBI.DistributorSeed, stellarSeedsOfSBI.IssuerSeed, "SBIT", "100")
	// var messages = make(chan string, 4)
	// PrintBalencesOfSIDaccounts(stellarAddressesOfSBI)
	// PrintBalencesOfSIDaccounts(stellarAddressesOfJPM)
	// go transaction.WatchLiveActivityOf(stellarAddressesOfSBI.Distributor.Address(), messages)
	// go transaction.WatchLiveActivityOf(stellarAddressesOfJPM.Distributor.Address(), messages)
	// sendAssetFromAtoB(stellarAddressesOfJPM.Distributor, stellarAddressesOfSBI.Distributor, stellarSeedsOfJPM.DistributorSeed, JpmtAsset, "22")
	// println(<-messages)
	// println(<-messages)
	// PrintBalencesOfSIDaccounts(stellarAddressesOfSBI)
	// PrintBalencesOfSIDaccounts(stellarAddressesOfJPM)
}
