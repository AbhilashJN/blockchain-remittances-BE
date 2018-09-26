package main

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Registration successful: %s => %+v", r.FormValue("PhoneNumber"), cd)
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
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		senderInfo, err := db.ReadCustomerDetailsFromCommonCustomersDB(r.FormValue("senderPhone"))
		if err != nil {
			fmt.Fprintf(w, "reading failed: %v", err)
			return
		}
		senderBankStellarAdresses, err := db.ReadStellarAddressesOfBank(senderInfo.BankName)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}
		senderBankSeed := senderBankStellarAdresses.DistributorSeed
		senderBankAddress, err := getKeyPair(senderBankStellarAdresses.DistributorSeed)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}

		receiverInfo, err := db.ReadCustomerDetailsFromCommonCustomersDB(r.FormValue("receiverPhone"))
		if err != nil {
			fmt.Fprintf(w, "reading failed: %v", err)
			return
		}

		receiverBankStellarAdresses, err := db.ReadStellarAddressesOfBank(receiverInfo.BankName)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}

		receiverBankAddress, err := getKeyPair(receiverBankStellarAdresses.DistributorSeed)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}

		sendAssetFromAtoB(senderBankAddress, receiverBankAddress, senderBankSeed, buildAsset(senderBankAddress, senderInfo.BankName+"T"), r.FormValue("Amount"))
	}
}

//StartServer starts the server
func StartServer() {
	http.HandleFunc("/ping", pong)
	http.HandleFunc("/registration", registration)
	http.HandleFunc("/getReceiverInfo", getReceiverInfo)
	http.HandleFunc("/sendPayment", sendPayment)
	http.ListenAndServe("localhost:8080", nil)
}
