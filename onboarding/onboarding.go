package onboarding

import (
	"fmt"
	"log"

	"github.com/AbhilashJN/blockchain-remittances-BE/account"
	"github.com/AbhilashJN/blockchain-remittances-BE/transaction"
	"github.com/AbhilashJN/blockchain-remittances-BE/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
)

// createNewAccFromSourceAcc returns
func createNewAccFromSourceAcc(sourceAcc *keypair.Full, initBalance string) (*keypair.Full, error) {
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
		return nil, err
	}

	txe, err := tx.Sign(sourceAccSeed)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = horizon.DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		return nil, err
	}

	return newAccountKeyPair, nil
}

// CreateDistributionAccount returns
func createDistributionAccount(sourceAccount *keypair.Full, initBalance string) (*keypair.Full, error) {
	return createNewAccFromSourceAcc(sourceAccount, initBalance)
}

// CreateIssuingAccount returns
func createIssuingAccount(sourceAccount *keypair.Full, initBalance string) (*keypair.Full, error) {
	return createNewAccFromSourceAcc(sourceAccount, initBalance)
}

// CreateSourceAccount returns
func createSourceAccount() (*keypair.Full, error) {
	keyPair, err := account.CreateTestAccount()
	if err != nil {
		return nil, err
	}
	return keyPair, nil
}

// CreateTrust returns
func CreateTrust(receiver *keypair.Full, sender *keypair.Full, assetCode string) error {

	tx, err := build.Transaction(
		build.SourceAccount{AddressOrSeed: receiver.Address()},
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.TestNetwork,
		build.Trust(assetCode, sender.Address()),
	)

	if err != nil {
		fmt.Println(err)
		return err
	}

	txe, err := tx.Sign(receiver.Seed())
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
func IssueToDistribAccount(issuerSeed, issuerAddress, receiverAddress, assetCode, amount string) error {

	paymentTx, err := build.Transaction(
		build.SourceAccount{AddressOrSeed: issuerAddress},
		build.TestNetwork,
		build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
		build.Payment(
			build.Destination{AddressOrSeed: receiverAddress},
			build.CreditAmount{Code: assetCode, Issuer: issuerAddress, Amount: amount},
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
func OnboardBank(assetCode, distAccInitBalance, issuingAccMinBal, distributingAccMinBal string) error {

	sourceAccKeyPair, err := createSourceAccount()
	if err != nil {
		return err
	}

	fmt.Printf("sourceAccKeyPair:\n %s\n%s\n", sourceAccKeyPair.Seed(), sourceAccKeyPair.Address())

	issuerKeyPair, err := createIssuingAccount(sourceAccKeyPair, issuingAccMinBal)
	if err != nil {
		return err
	}

	fmt.Printf("issuerKeyPair:\n %s\n%s\n", issuerKeyPair.Seed(), issuerKeyPair.Address())

	distributorKeyPair, err := createDistributionAccount(sourceAccKeyPair, distributingAccMinBal)
	if err != nil {
		return err
	}

	fmt.Printf("distributorKeyPair:\n %s\n%s\n", distributorKeyPair.Seed(), distributorKeyPair.Address())

	err = CreateTrust(distributorKeyPair, issuerKeyPair, assetCode)
	if err != nil {
		return err
	}

	_, err = transaction.SendPaymentTransaction(distAccInitBalance, issuerKeyPair.Address(), distributorKeyPair.Address(), issuerKeyPair.Seed(), "initpump", utils.BuildAsset(issuerKeyPair.Address(), assetCode))
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
