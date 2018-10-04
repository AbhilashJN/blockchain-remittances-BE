package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/AbhilashJN/blockchain-remittances-BE/data"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	"github.com/AbhilashJN/blockchain-remittances-BE/account"
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

func buildAsset(issuerAddress, assetCode string) build.Asset {
	asset := build.CreditAsset(assetCode, issuerAddress)
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

func getRecipient() (keypair.KP, string) {
	personB := account.GetCreatedAccounts().PersonB
	recipientSeed := personB.Seed
	recipientKP, err := keypair.Parse(recipientSeed)
	if err != nil {
		log.Fatal(err)
	}
	return recipientKP, recipientSeed
}

func sendPaymentTransaction(amount, senderStellarAddress, receiverStellarAddress, senderStellarSeed, memo string, asset build.Asset) (*horizon.TransactionSuccess, error) {
	paymentTx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{AddressOrSeed: senderStellarAddress},
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.MemoText{Value: memo},
		build.Payment(
			build.Destination{AddressOrSeed: receiverStellarAddress},
			build.CreditAmount{Code: asset.Code, Issuer: asset.Issuer, Amount: amount},
		),
	)
	if err != nil {
		return nil, err
	}

	paymentTxe, err := paymentTx.Sign(senderStellarSeed)
	if err != nil {
		return nil, err
	}

	paymentTxeB64, err := paymentTxe.Base64()
	if err != nil {
		return nil, err
	}

	resp, err := horizon.DefaultTestNetClient.SubmitTransaction(paymentTxeB64)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func createAndSendCustomTokenFromAtoB(issuerKP, recipientKP keypair.KP, issuerSeed, recipientSeed string, customAsset build.Asset, assetAmout string) {

	fmt.Println("_____________________________________before_____________________________________")
	account.PrintAccountDetails(issuerKP.Address())
	account.PrintAccountDetails(recipientKP.Address())
	fmt.Println("_____________________________________before_____________________________________")

	trustAsset(recipientKP, recipientSeed, customAsset)

	// sendAssetFromAtoB(issuerKP, recipientKP, issuerSeed, customAsset, assetAmout)

	fmt.Println("_____________________________________after_______________________________________")
	account.PrintAccountDetails(issuerKP.Address())
	account.PrintAccountDetails(recipientKP.Address())
	fmt.Println("_____________________________________after_______________________________________")
}

// PrintBalencesOfSIDaccounts returns
func PrintBalencesOfSIDaccounts(stellarAdresses *data.StellarAddresses) {
	fmt.Println("SOURCE ACCOUNT START------------------------------------------------\n-")
	account.PrintAccountDetails(stellarAdresses.Source)
	fmt.Println("SOURCE ACCOUNT END------------------------------------------------\n-")
	fmt.Println("ISSUER ACCOUNT START:------------------------------------------------\n-")
	account.PrintAccountDetails(stellarAdresses.Issuer)
	fmt.Println("ISSUER ACCOUNT END------------------------------------------------\n-")
	fmt.Println("DISTRIBUTION ACCOUNT START:------------------------------------------------\n-")
	account.PrintAccountDetails(stellarAdresses.Distributor)
	fmt.Println("DISTRIBUTION ACCOUNT END------------------------------------------------\n-")
}

var filePath = flag.String("configFile", "./SBIconfig.yml", "config filepath")

type Keys struct {
	Source, Issuer, Distributor string
}

type DBconnectionParams struct {
	Host, Port, User, Password, DbName string
}

type BankConfig struct {
	Name             string
	Port             string
	StellarSeeds     Keys
	StellarAddresses Keys
	DBconnectionParams
	DB *gorm.DB
}

var bankConfig BankConfig

func readConfig(configFilePath string) {
	viper.SetConfigFile(configFilePath)

	// Searches for config file in given paths and read it
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Confirm which config file is used
	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())

	if err := viper.Unmarshal(&bankConfig); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

func init() {
	flag.Parse()
	readConfig(*filePath)
	spew.Dump(bankConfig)
}

func main() {
	// flag.Parse()
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

	// stellarSeedsOfSBI, err := db.ReadStellarSeedsOfBank("SBI")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // fmt.Printf("%+v \n", stellarSeedsOfSBI)

	// stellarSeedsOfJPM, err := db.ReadStellarSeedsOfBank("JPM")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%+v \n", stellarSeedsOfJPM)
	// stellarAddressesOfSBI, err := GetStellarAddressesOfBank(&db.StellarSeeds{
	// 	SourceSeed:      "SBNSSMFUYGUIPXMOEKSAGD524THQKE6S5NUEGZVZ3LNB422G427DNAIL",
	// 	IssuerSeed:      "SBZK7WTUPFIY55HRRW4SYH2KFZVMLH7STS2PTQHOW4RAF7UWRB74BGKM",
	// 	DistributorSeed: "SDP3446MMEKWDSUKJ65GASOBNL6L42UVBPAXKKGOMCSVARB3WMLHJVQY",
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// stellarAddressesOfJPM, err := GetStellarAddressesOfBank(&db.StellarSeeds{
	// 	SourceSeed:      "SB7SZBJ5BUYWQGQZ3TVSTDN7FY7F6ERPFI22TBJRDBYHOKUDRVLG4KFW",
	// 	IssuerSeed:      "SBJW3HMQG4AUCFO63YGNCBJRQAODMGOOQBLQT74TBSUYHYPDU56WZB3V",
	// 	DistributorSeed: "SDZ2T2L4LT76MIPUIA5LJLYGPDSF4CWIWVID5U32H6N62N5OEZ2VSTBX",
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("stellarAddressesOfSBI.Source.Address: %q\n", stellarAddressesOfSBI.Source.Address())
	// fmt.Printf("stellarAddressesOfSBI.Issuer.Address: %q\n", stellarAddressesOfSBI.Issuer.Address())
	// fmt.Printf("stellarAddressesOfSBI.Distributor.Address: %q\n", stellarAddressesOfSBI.Distributor.Address())
	// fmt.Printf("stellarAddressesOfJPM.Source.Address: %q\n", stellarAddressesOfJPM.Source.Address())
	// fmt.Printf("stellarAddressesOfJPM.Issuer.Address: %q\n", stellarAddressesOfJPM.Issuer.Address())
	// fmt.Printf("stellarAddressesOfJPM.Distributor.Address: %q\n", stellarAddressesOfJPM.Distributor.Address())

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

	// stellarSeeds, err := db.ReadStellarSeedsOfBank("JPM")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// spew.Dump(stellarSeeds)

	// stellarAddresses, err := utils.GetStellarAddressesOfBank(stellarSeeds)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// spew.Dump(stellarAddresses)

	// bank := &bank.Bank{Name: bankName, StellarSeeds: stellarSeeds, StellarAddresses: stellarAddresses}

	// os.Setenv("StellarIssuerSeed", stellarSeeds.Issuer)
	// os.Setenv("StellarDistributorSeed", stellarSeeds.Distributor)
	// os.Setenv("StellarSourceSeed", stellarSeeds.Source)

	// os.Setenv("StellarIssuerAddress", stellarAddresses.Issuer)
	// os.Setenv("StellarDistributorAddress", stellarAddresses.Distributor)
	// os.Setenv("StellarSourceAddress", stellarAddresses.Source)

	// err = setup.IssueToDistribAccount(bank.StellarSeeds.Issuer, bank.StellarAddresses.Issuer, bank.StellarAddresses.Distributor, bankName+"T", "100")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf("Server for %q,\nAccount Address: %q \n", bank.Name, bank.StellarAddresses.Distributor)

	dbConn := bankConfig.DBconnectionParams
	dbConnParam := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", dbConn.Host, dbConn.Port, dbConn.User, dbConn.DbName, dbConn.Password)

	db, err := gorm.Open("postgres", dbConnParam)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	bankConfig.DB = db

	// go ListenForPayments(bankConfig)
	StartServer(bankConfig)
}
