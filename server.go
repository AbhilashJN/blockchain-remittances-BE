package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/AbhilashJN/blockchain-remittances-BE/bank"
	"github.com/AbhilashJN/blockchain-remittances-BE/utils"
	"github.com/stellar/go/clients/horizon"

	"github.com/AbhilashJN/blockchain-remittances-BE/data"

	"github.com/AbhilashJN/blockchain-remittances-BE/db"
)

func pong(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintln(w, "pong")
	}
}

// ListenForPayments returns
func ListenForPayments(bankConfig BankConfig) {
	ctx := context.Background()

	cursor := horizon.Cursor("now")

	fmt.Println("Waiting for a payment...")

	err := horizon.DefaultTestNetClient.StreamTransactions(ctx, bank.StellarAddresses.Distributor, &cursor,
		func(transaction horizon.Transaction) {
			if err := handleTransaction(bank, transaction); err != nil {
				log.Printf("In callback of StreamTransactions: %s", err.Error())
			}
		},
	)

	if err != nil {
		fmt.Printf("shit happened")
		panic(err)
	}

}

func handleTransaction(bank *bank.Bank, transaction horizon.Transaction) error {
	if bank.StellarAddresses.Distributor == transaction.Account {
		return nil
	}

	if bank.StellarAddresses.Issuer == transaction.Account {
		fmt.Println("transaction from issuer account")
		return nil
	}

	fmt.Println("\n\nReceived a transaction from stellar network..")

	txe, err := utils.DecodeTransactionEnvelope(transaction.EnvelopeXdr)
	if err != nil {
		return err
	}
	// spew.Dump(transaction) //pretty print function
	fields := strings.Split(transaction.Memo, ";")
	customerAccountIDtoCredit, senderAccountID, senderName := fields[0], fields[1], fields[2]
	operation := txe.Tx.Operations[0].Body.PaymentOp
	amount := float64(operation.Amount) / 1e7 // TODO: Verify the validity of this
	assetInfo, ok := operation.Asset.GetAlphaNum4()
	if !ok {
		return errors.New("GetAlphaNum4() failed: Could not extract alpha4 asset from the envelope operation")
	}

	transactionDetails := &data.TransactionDetails{TransactionType: "credit", From: senderAccountID, Amount: amount, TransactionID: transaction.ID}

	fmt.Printf("Asset code: %q\n", assetInfo.AssetCode)
	fmt.Printf("Amount: %f\n", transactionDetails.Amount)
	fmt.Printf("From bank account: %q, name: %q \n", transactionDetails.From, senderName)
	fmt.Printf("Bank account to credit: %q\n", customerAccountIDtoCredit)
	updatedCustomerAccountInfo, updatedBankPoolAccountInfo, err := bank.UpdateCustomerBankAccountBalence(transactionDetails, customerAccountIDtoCredit)
	if err != nil {
		return err
	}
	utils.LogAccountDetails(updatedCustomerAccountInfo, updatedBankPoolAccountInfo)
	return nil
}

func getReceiverInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err\n: %v", err)
			return
		}
		receiverCustomerBankAccountDetails, err := db.ReadCustomerDetailsFromCustomerPoolDB(r.FormValue("PhoneNumber"))
		if err != nil {
			fmt.Fprintf(w, "reading failed: %v", err)
			return
		}

		fmt.Printf("%+v\n", receiverCustomerBankAccountDetails)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsEncoder := json.NewEncoder(w)
		err = jsEncoder.Encode(receiverCustomerBankAccountDetails)
		if err != nil {
			fmt.Fprintf(w, "jsEncoder.Encode(receiverCustomerBankAccountDetails) failed:\n errorL %v", err)
			return
		}
	}
}

func sendPayment(w http.ResponseWriter, r *http.Request, bank *bank.Bank) {

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		senderName := r.FormValue("senderName")
		senderBankAccountID := r.FormValue("senderBankAccountID")
		receiverBankAccountID := r.FormValue("receiverBankAccountID")
		receiverBankStellarDistributorAddress := r.FormValue("receiverBankStellarDistributorAddress")
		amountToCredit := r.FormValue("Amount")
		amountInFloat, err := strconv.ParseFloat(amountToCredit, 64)
		if err != nil {
			fmt.Fprintf(w, "strconv.ParseFloat(amountToCredit, 64) failed\n: %v", err)
			return
		}

		resp, err := sendPaymentTransaction(amountToCredit, bank.StellarAddresses.Distributor, receiverBankStellarDistributorAddress,
			bank.StellarSeeds.Distributor, fmt.Sprintf("%s;%s;%s", receiverBankAccountID, senderBankAccountID, senderName),
			buildAsset(bank.StellarAddresses.Issuer, bank.Name+"T"))

		if err != nil {
			fmt.Fprintf(w, "error in send payment transaction: %v", err)
			return
		}

		fmt.Printf("Successful payment transaction by %q on the stellar network\n:", bank.Name)

		transactionDetails := &data.TransactionDetails{TransactionType: "debit", To: receiverBankAccountID, Amount: amountInFloat, TransactionID: resp.Hash}

		updatedCustomerAccountInfo, updatedBankPoolAccountInfo, err := bank.UpdateCustomerBankAccountBalence(transactionDetails, senderBankAccountID)
		if err != nil {
			fmt.Fprintf(w, "db.UpdateCustomerBankAccountBalence(transactionDetails, bankName, customerAccountIDtoCredit) failed:\n %v", err.Error())
			return
		}

		fmt.Println("\n\nSender customer bank account details after succesful transaction")
		utils.LogAccountDetails(updatedCustomerAccountInfo, updatedBankPoolAccountInfo)
		fmt.Fprintf(w, "success")
	}
}

func getCustomerAccountDetails(w http.ResponseWriter, r *http.Request, bank *bank.Bank) {
	if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		customerBankAccountDetails, err := db.ReadCustomerBankAccountDetails(bank.Name, r.FormValue("BankAccountID"))
		if err != nil {
			fmt.Fprintf(w, "reading failed: %v", err)
			return
		}
		fmt.Println(r.FormValue("BankAccountID"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsEncoder := json.NewEncoder(w)
		err = jsEncoder.Encode(customerBankAccountDetails)
		if err != nil {
			fmt.Fprintf(w, "jsEncoder.Encode(customerBankAccountDetails) failed:\n error %v", err)
			return
		}
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *bank.Bank), bank *bank.Bank) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, bank)
	}
}

//StartServer starts the server
func StartServer(bankConfig BankConfig) {
	http.HandleFunc("/ping", pong)
	http.HandleFunc("/getReceiverInfo", getReceiverInfo)
	http.HandleFunc("/sendPayment", makeHandler(sendPayment, bankConfig))
	http.HandleFunc("/getCustomerAccountDetails", makeHandler(getCustomerAccountDetails, bankConfig))
	fmt.Println("\n\nserver is starting...")
	err := http.ListenAndServe(fmt.Sprintf("localhost:%s", bankConfig.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
