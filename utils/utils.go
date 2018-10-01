package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/AbhilashJN/blockchain-remittances-BE/data"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/xdr"
)

// DecodeTransactionEnvelope returns
func DecodeTransactionEnvelope(data string) (xdr.TransactionEnvelope, error) {

	rawr := strings.NewReader(data)
	b64r := base64.NewDecoder(base64.StdEncoding, rawr)

	var tx xdr.TransactionEnvelope
	bytesRead, err := xdr.Unmarshal(b64r, &tx)

	if err != nil {
		return tx, err
	}

	fmt.Printf("Successful decoding of transaction envelope. Read %d bytes\n", bytesRead)
	return tx, nil
}

// LogAccountDetails returns
func LogAccountDetails(updatedCustomerAccountInfo, updatedBankPoolAccountInfo *data.CustomerBankAccountDetails) {
	fmt.Printf("Account holder name: %q\n", updatedCustomerAccountInfo.Name)
	fmt.Printf("Account holder balance: %f\n", updatedCustomerAccountInfo.Balance)
	logTransactionHistory(updatedCustomerAccountInfo.Transactions)
	fmt.Println("\nBank pool account detils")
	fmt.Printf("Balance: %f\n", updatedBankPoolAccountInfo.Balance)
	logTransactionHistory(updatedBankPoolAccountInfo.Transactions)

}

func logTransactionHistory(transactionHistory []data.TransactionDetails) {
	fmt.Println("--------------------------------------------Transaction history start----------------------------------------------")
	for _, tx := range transactionHistory {
		fmt.Println("------------------------------------------------------------------------------------------")
		if tx.TransactionType == "credit" {
			fmt.Printf("TransactionID: %q\nTransactionType: %q\nFrom: %q\nAmount: %f\n", tx.TransactionID, tx.TransactionType, tx.From, tx.Amount)
		} else {
			fmt.Printf("TransactionID: %q\nTransactionType: %q\nTo: %q\nAmount: %f\n", tx.TransactionID, tx.TransactionType, tx.To, tx.Amount)
		}
		fmt.Println("------------------------------------------------------------------------------------------")
	}
	fmt.Println("--------------------------------------------Transaction history end----------------------------------------------")
}

// EnvVarNotFoundError returns
func EnvVarNotFoundError(got string) error {
	return fmt.Errorf("environment variable %q not found", got)
}

// GetStellarAddressesOfBank returns
func GetStellarAddressesOfBank(stellarSeedsOfBank *data.StellarSeeds) (*data.StellarAddresses, error) {
	SourceKeyPair, err := getKeyPair(stellarSeedsOfBank.Source)
	if err != nil {
		return nil, err
	}
	IssuerKeyPair, err := getKeyPair(stellarSeedsOfBank.Issuer)
	if err != nil {
		return nil, err
	}
	DistributorKeyPair, err := getKeyPair(stellarSeedsOfBank.Distributor)
	if err != nil {
		return nil, err
	}
	return &data.StellarAddresses{Keys: data.Keys{Source: SourceKeyPair.Address(), Issuer: IssuerKeyPair.Address(), Distributor: DistributorKeyPair.Address()}}, nil
}

func getKeyPair(seed string) (keypair.KP, error) {
	accKeyPair, err := keypair.Parse(seed)
	if err != nil {
		return nil, err
	}
	return accKeyPair, nil
}

func GobEncode(structValue interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(structValue)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GobDecode(data []byte, target data.Data) (data.Data, error) {
	// fmt.Printf("In gobDecode: %s", data)
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(target)
	if err != nil {
		return nil, err
	}
	return target, nil
}
