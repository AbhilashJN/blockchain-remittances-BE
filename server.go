package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
		err := db.WriteCustomerDetailsToCommonCustomersDB(r.FormValue("PhoneNumber"), &db.CustomerDetails{
			CustomerName:  r.FormValue("CustomerName"),
			BankName:      r.FormValue("BankName"),
			BankAccountID: r.FormValue("BankAccountID"),
		})
		if err != nil {
			fmt.Fprintf(w, "registration failed: %v", err)
			return
		}
		cd, err := db.ReadCustomerDetailsFromCommonCustomersDB(r.FormValue("PhoneNumber"))
		if err != nil {
			fmt.Fprintf(w, "registration failed: %v", err)
			return
		}
		err = db.WriteCustomerBankAccountDetails(cd.BankName, cd.BankAccountID, &db.CustomerBankAccountDetails{Name: cd.CustomerName, Balance: 1000.0, Transactions: []db.TransactionDetails{}})
		if err != nil {
			fmt.Fprintf(w, "registration failed: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(*cd)

	}
}

func getReceiverInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		cd, err := db.ReadCustomerDetailsFromCommonCustomersDB(r.FormValue("PhoneNumber"))
		if err != nil {
			fmt.Fprintf(w, "reading failed: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(*cd)
	}
}

func sendPayment(w http.ResponseWriter, r *http.Request) {
	senderBankStellarSeeds, err := db.ReadStellarSeedsOfBank(*bankNameFlag)
	if err != nil {
		fmt.Fprintf(w, "ReadStellarSeedsOfBank failed: %v", err)
		return
	}

	senderBankStellarAddressKP, err := GetSIDkeyPairsOfBank(senderBankStellarSeeds)
	if err != nil {
		fmt.Fprintf(w, "GetSIDkeyPairsOfBank failed: %v", err)
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		senderName := r.FormValue("senderName")
		senderBankAccountID := r.FormValue("senderBankAccountID")
		receiverBank := r.FormValue("receiverBankName")
		receiverBankAccountID := r.FormValue("receiverBankAccountID")
		amountToCredit := r.FormValue("Amount")
		amountInFloat, err := strconv.ParseFloat(amountToCredit, 64)
		if err != nil {
			fmt.Fprintf(w, "strconv.ParseFloat(amountToCredit, 64) failed: %v", err)
			return
		}

		receiverBankStellarSeeds, err := db.ReadStellarSeedsOfBank(receiverBank)
		if err != nil {
			fmt.Fprintf(w, "db.ReadStellarSeedsOfBank(receiverBank) failed: %v", err)
			return
		}

		receiverBankStellarAddressKP, err := GetSIDkeyPairsOfBank(receiverBankStellarSeeds)
		if err != nil {
			fmt.Fprintf(w, "GetSIDkeyPairsOfBank(receiverBankStellarSeeds) failederror: %v", err)
			return
		}

		resp, err := sendAssetFromAtoB(senderBankStellarAddressKP.Distributor, receiverBankStellarAddressKP.Distributor,
			senderBankStellarSeeds.DistributorSeed, buildAsset(senderBankStellarAddressKP.Issuer, *bankNameFlag+"T"),
			amountToCredit, fmt.Sprintf("%s;%s;%s", receiverBankAccountID, senderBankAccountID, senderName))

		if err != nil {
			fmt.Fprintf(w, "error in send payment transaction: %v", err)
			return
		}

		fmt.Printf("Successful payment transaction by %q to %q on the stellar network\n:", *bankNameFlag, receiverBank)
		// spew.Dump(resp)
		// fmt.Println("Ledger:", resp.Ledger)
		// fmt.Println("Hash:", resp.Hash)
		// txe, err := utils.DecodeTransactionEnvelope(resp.Env)
		// if err != nil {
		// 	fmt.Fprintf(w, "utils.DecodeTransactionEnvelope(resp.Env) failed: %v", err)
		// 	return
		// }

		transactionDetails := &db.TransactionDetails{TransactionType: "debit", To: receiverBankAccountID, Amount: amountInFloat, TransactionID: resp.Hash}

		updatedCustomerAccountInfo, updatedBankPoolAccountInfo, err := db.UpdateCustomerBankAccountBalence(transactionDetails, *bankNameFlag, senderBankAccountID)
		if err != nil {
			fmt.Fprintf(w, "db.UpdateCustomerBankAccountBalence(transactionDetails, bankName, customerAccountIDtoCredit) failed: %s", err.Error())
			return
		}
		// spew.Dump(updatedAccountDetails)
		fmt.Println("\n\nSender customer bank account details after succesful transaction")
		fmt.Printf("Account holder name: %q\n", updatedCustomerAccountInfo.Name)
		fmt.Printf("Account holder balance: %f\n", updatedCustomerAccountInfo.Balance)
		fmt.Println("--------------------------------------------Transaction history----------------------------------------------")
		for _, tx := range updatedCustomerAccountInfo.Transactions {
			fmt.Printf("TransactionID: %q\nTransactionType: %q\nTo: %q\nAmount: %f\n", tx.TransactionID, tx.TransactionType, tx.To, tx.Amount)
			fmt.Println("------------------------------------------------------------------------------------------")
		}
		fmt.Println("--------------------------------------------Transaction history----------------------------------------------")

		fmt.Println("\n\nBank pool account detils")
		// fmt.Printf("updatedBankPoolAccountInfo : %T\n\n", updatedBankPoolAccountInfo)
		fmt.Printf("Balance: %f\n", updatedBankPoolAccountInfo.Balance)
		fmt.Println("--------------------------------------------Transaction history----------------------------------------------")
		for _, tx := range updatedBankPoolAccountInfo.Transactions {
			fmt.Printf("TransactionID: %q\nTransactionType: %q\nFrom: %q\n, Amount: %f\n", tx.TransactionID, tx.TransactionType, tx.From, tx.Amount)
		}
		fmt.Println("--------------------------------------------Transaction history----------------------------------------------")

		fmt.Fprintf(w, "success")
	}
}

func getTransactionDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		cd, err := db.ReadCustomerBankAccountDetails(*bankNameFlag, r.FormValue("BankAccountID"))
		if err != nil {
			fmt.Fprintf(w, "reading failed: %v", err)
			return
		}
		fmt.Println(r.FormValue("BankAccountID"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(*cd)
	}
}

//StartServer starts the server
func StartServer() {
	http.HandleFunc("/ping", pong)
	http.HandleFunc("/registration", registration)
	http.HandleFunc("/getReceiverInfo", getReceiverInfo)
	http.HandleFunc("/sendPayment", sendPayment)
	http.HandleFunc("/getTransactionDetails", getTransactionDetails)
	fmt.Println("\n\nserver is starting...")
	err := http.ListenAndServe(fmt.Sprintf("localhost:%s", *portNumFlag), nil)
	if err != nil {
		log.Fatal(err)
	}
}
