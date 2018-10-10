package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/xdr"
)

type PaymentInfo struct {
	AssetCode         string
	Amount            float64
	SenderAccountID   string
	ReceiverAccountID string
	SenderName        string
	TxID              string
}

// var exchangeRates = map[string]type

func convertSenderAssetAmountToBankNativeAssetAmount(senderAssetCode string, amountInSenderCurrency float64) float64 {
	var amountInBankNativeCurrency float64
	switch senderAssetCode {
	case "JPMT":
		amountInBankNativeCurrency = amountInSenderCurrency * 80
	case "SBIT":
		amountInBankNativeCurrency = amountInSenderCurrency / 80
	}
	return amountInBankNativeCurrency
}

// DecodeTransactionEnvelope returns
func DecodeTransactionEnvelope(transaction horizon.Transaction) (*PaymentInfo, error) {

	rawr := strings.NewReader(transaction.EnvelopeXdr)
	b64r := base64.NewDecoder(base64.StdEncoding, rawr)

	var txEnv xdr.TransactionEnvelope
	bytesRead, err := xdr.Unmarshal(b64r, &txEnv)

	if err != nil {
		return nil, err
	}

	// spew.Dump(transaction) //pretty print function
	fields := strings.Split(transaction.Memo, ";")
	receiverAccountID, senderAccountID, senderName := fields[0], fields[1], fields[2]
	operation := txEnv.Tx.Operations[0].Body.PaymentOp
	amountInSenderCurrency := float64(operation.Amount) / 1e7 // TODO: Verify the validity of this
	senderAssetInfo, ok := operation.Asset.GetAlphaNum4()
	if !ok {
		return nil, errors.New("GetAlphaNum4() failed: Could not extract alpha4 asset from the envelope operation")
	}

	senderAssetCode := string(senderAssetInfo.AssetCode[:4])

	log.Printf("Asset code: %q\n", senderAssetInfo.AssetCode)
	log.Printf("Amount in sender currency: %f\n", amountInSenderCurrency)
	log.Printf("From bank account: %q, name: %q \n", senderAccountID, senderName)
	log.Printf("Bank account to credit: %q\n", receiverAccountID)
	fmt.Printf("Successful decoding of transaction envelope. Read %d bytes\n", bytesRead)

	amountInBankNativeCurrency := convertSenderAssetAmountToBankNativeAssetAmount(senderAssetCode, amountInSenderCurrency)

	return &PaymentInfo{
		TxID: transaction.ID, AssetCode: senderAssetCode, Amount: amountInBankNativeCurrency, SenderAccountID: senderAccountID, SenderName: senderName, ReceiverAccountID: receiverAccountID,
	}, nil
}

// LogAccountDetails returns
// func LogAccountDetails(updatedCustomerAccountInfo, updatedBankPoolAccountInfo *data.CustomerBankAccountDetails) {
// 	fmt.Printf("Account holder name: %q\n", updatedCustomerAccountInfo.Name)
// 	fmt.Printf("Account holder balance: %f\n", updatedCustomerAccountInfo.Balance)
// 	logTransactionHistory(updatedCustomerAccountInfo.Transactions)
// 	fmt.Println("\nBank pool account detils")
// 	fmt.Printf("Balance: %f\n", updatedBankPoolAccountInfo.Balance)
// 	logTransactionHistory(updatedBankPoolAccountInfo.Transactions)

// }

// func logTransactionHistory(transactionHistory []data.TransactionDetails) {
// 	fmt.Println("--------------------------------------------Transaction history start----------------------------------------------")
// 	for _, tx := range transactionHistory {
// 		fmt.Println("------------------------------------------------------------------------------------------")
// 		if tx.TransactionType == "credit" {
// 			fmt.Printf("TransactionID: %q\nTransactionType: %q\nFrom: %q\nAmount: %f\n", tx.TransactionID, tx.TransactionType, tx.From, tx.Amount)
// 		} else {
// 			fmt.Printf("TransactionID: %q\nTransactionType: %q\nTo: %q\nAmount: %f\n", tx.TransactionID, tx.TransactionType, tx.To, tx.Amount)
// 		}
// 		fmt.Println("------------------------------------------------------------------------------------------")
// 	}
// 	fmt.Println("--------------------------------------------Transaction history end----------------------------------------------")
// }

// EnvVarNotFoundError returns
// func EnvVarNotFoundError(got string) {
// 	log.Fatal(fmt.Errorf("environment variable %q not found", got))
// }

// GetStellarAddressesOfBank returns
func GetStellarAddressesOfBank(sourceSeed, issuerSeed, distributorSeed string) (string, string, string, error) {
	sourceAddress, err := getAddressFromSeed(sourceSeed)
	if err != nil {
		return "", "", "", err
	}
	issuerAddress, err := getAddressFromSeed(issuerSeed)
	if err != nil {
		return "", "", "", err
	}
	distributorAddress, err := getAddressFromSeed(distributorSeed)
	if err != nil {
		return "", "", "", err
	}
	return sourceAddress, issuerAddress, distributorAddress, err
}

func getAddressFromSeed(seed string) (string, error) {
	accKeyPair, err := keypair.Parse(seed)
	if err != nil {
		return "", err
	}
	return accKeyPair.Address(), nil
}

// BuildAsset returns
func BuildAsset(issuerAddress, assetCode string) build.Asset {
	asset := build.CreditAsset(assetCode, issuerAddress)
	return asset
}

// // GobEncode returns
// func GobEncode(structValue interface{}) ([]byte, error) {
// 	buf := new(bytes.Buffer)
// 	enc := gob.NewEncoder(buf)
// 	err := enc.Encode(structValue)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

// // GobDecode returns
// func GobDecode(data []byte, target data.Data) (data.Data, error) {
// 	// fmt.Printf("In gobDecode: %s", data)
// 	buf := bytes.NewBuffer(data)
// 	dec := gob.NewDecoder(buf)
// 	err := dec.Decode(target)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return target, nil
// }
const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// CreateRandomString returns
func CreateRandomString() string {
	n := 50
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
