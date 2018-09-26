package bank

import (
	"fmt"
	"log"

	"github.com/AbhilashJN/blockchain-remittances-BE/account"
	"github.com/AbhilashJN/blockchain-remittances-BE/db"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
)

// SIDKeyPairs returns
type SIDKeyPairs struct {
	Source, Issuer, Distributor keypair.KP
}

// CreateNewAccFromSourceAcc returns
func CreateNewAccFromSourceAcc(sourceAcc *keypair.Full, initBalance string) (string, error) {
	sourceAccSeed := sourceAcc.Seed()
	newAccountKeyPair := account.MakePair()

	tx, err := build.Transaction(
		build.SourceAccount{AddressOrSeed: sourceAccSeed},
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.TestNetwork,
		build.CreateAccount(
			build.Destination{AddressOrSeed: newAccountKeyPair.Address()},
			build.NativeAmount{Amount: initBalance},
		),
	)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	txe, err := tx.Sign(sourceAccSeed)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	_, err = horizon.DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		return "", err
	}

	return newAccountKeyPair.Seed(), nil
}

// CreateDistributionAccount returns
func CreateDistributionAccount(sourceAccount *keypair.Full, initBalance string) (string, error) {
	return CreateNewAccFromSourceAcc(sourceAccount, initBalance)
}

// CreateIssuingAccount returns
func CreateIssuingAccount(sourceAccount *keypair.Full, initBalance string) (string, error) {
	return CreateNewAccFromSourceAcc(sourceAccount, initBalance)
}

// CreateSourceAccount returns
func CreateSourceAccount() (*keypair.Full, error) {
	keyPair, err := account.CreateTestAccount()
	if err != nil {
		return nil, err
	}
	return keyPair, nil
}

// CreateTrust returns
func CreateTrust(receiverSeed, senderSeed string, assetCode string) error {
	issuer, err := keypair.Parse(senderSeed)
	if err != nil {
		log.Fatal(err)
	}

	receiver, err := keypair.Parse(receiverSeed)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := build.Transaction(
		build.SourceAccount{AddressOrSeed: receiver.Address()},
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.TestNetwork,
		build.Trust(assetCode, issuer.Address()),
	)

	if err != nil {
		fmt.Println(err)
		return err
	}

	txe, err := tx.Sign(receiverSeed)
	if err != nil {
		fmt.Println(err)
		return err
	}

	txeB64, err := txe.Base64()
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = horizon.DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// IssueToDistribAccount returns
func IssueToDistribAccount(receiverSeed, issuerSeed string, assetCode, amount string) error {
	issuer, err := keypair.Parse(issuerSeed)
	if err != nil {
		log.Fatal(err)
	}

	receiver, err := keypair.Parse(receiverSeed)
	if err != nil {
		log.Fatal(err)
	}

	paymentTx, err := build.Transaction(
		build.SourceAccount{AddressOrSeed: issuer.Address()},
		build.TestNetwork,
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.Payment(
			build.Destination{AddressOrSeed: receiver.Address()},
			build.CreditAmount{Code: assetCode, Issuer: issuer.Address(), Amount: amount},
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

	return nil
}

// OnboardBank returns
func OnboardBank(bankName string, assetCode string) error {
	err := db.CreateDBForBank(bankName)
	if err != nil {
		return err
	}

	sourceAccKeyPair, err := CreateSourceAccount()
	if err != nil {
		return err
	}

	issuerSeed, err := CreateIssuingAccount(sourceAccKeyPair, "100")
	if err != nil {
		return err
	}

	distributorSeed, err := CreateDistributionAccount(sourceAccKeyPair, "100")
	if err != nil {
		return err
	}

	err = CreateTrust(distributorSeed, issuerSeed, assetCode)
	if err != nil {
		return err
	}

	err = db.WriteStellarAddressesForBank(bankName, &db.StellarAddressesOfBank{
		SourceSeed:      sourceAccKeyPair.Seed(),
		IssuerSeed:      issuerSeed,
		DistributorSeed: distributorSeed,
	})
	if err != nil {
		return err
	}

	return nil
}
