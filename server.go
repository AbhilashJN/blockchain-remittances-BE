package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/AbhilashJN/blockchain-remittances-BE/bank"
	"github.com/AbhilashJN/blockchain-remittances-BE/utils"

	"github.com/AbhilashJN/blockchain-remittances-BE/data"

	"github.com/AbhilashJN/blockchain-remittances-BE/db"
)

func pong(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintln(w, "pong")
	}
}

func registration(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		err := db.WriteCustomerDetailsToCustomerPoolDB(r.FormValue("PhoneNumber"), &data.CustomerDetails{
			CustomerName:  r.FormValue("CustomerName"),
			BankName:      r.FormValue("BankName"),
			BankAccountID: r.FormValue("BankAccountID"),
		})
		if err != nil {
			fmt.Fprintf(w, "registration failed: %v", err)
			return
		}

		customerDetails, err := db.ReadCustomerDetailsFromCustomerPoolDB(r.FormValue("PhoneNumber"))
		if err != nil {
			fmt.Fprintf(w, "registration failed: %v", err)
			return
		}
		err = db.WriteCustomerBankAccountDetails(customerDetails.BankName, customerDetails.BankAccountID, &data.CustomerBankAccountDetails{Name: customerDetails.CustomerName, Balance: 1000.0})
		if err != nil {
			fmt.Fprintf(w, "registration failed: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		jsEncoder := json.NewEncoder(w)
		err = jsEncoder.Encode(customerDetails)
		if err != nil {
			fmt.Fprintf(w, "jsEncoder.Encode(customerDetails) failed:\n error %v", err)
			return
		}
	}
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

func getTransactionDetails(w http.ResponseWriter, r *http.Request, bank *bank.Bank) {
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
func StartServer(port string, bank *bank.Bank) {
	http.HandleFunc("/ping", pong)
	http.HandleFunc("/registration", registration)
	http.HandleFunc("/getReceiverInfo", getReceiverInfo)
	http.HandleFunc("/sendPayment", makeHandler(sendPayment, bank))
	http.HandleFunc("/getTransactionDetails", makeHandler(getTransactionDetails, bank))
	fmt.Println("\n\nserver is starting...")
	err := http.ListenAndServe(fmt.Sprintf("localhost:%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
