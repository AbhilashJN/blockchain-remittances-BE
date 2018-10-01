package data

import (
	"fmt"
)

// Data is
type Data interface {
	String()
}

// Keys implements data, contains
type Keys struct {
	Source, Issuer, Distributor string
}

// StellarSeeds is
type StellarSeeds struct {
	Keys
}

// StellarAddresses is
type StellarAddresses struct {
	Keys
}

func (seeds *StellarSeeds) String() {
	fmt.Printf("%+v\n", seeds)
}

func (addresses *StellarAddresses) String() {
	fmt.Printf("%+v\n", addresses)
}

//TransactionDetails contains
type TransactionDetails struct {
	From, To, TransactionID string
	TransactionType         string // debit or credit
	Amount                  float64
}

//CustomerBankAccountDetails implements data, contains
type CustomerBankAccountDetails struct {
	Name         string
	Balance      float64
	Transactions []TransactionDetails
}

// CustomerDetails implements data, contains
type CustomerDetails struct {
	CustomerName, BankName, BankAccountID, StellarDistributionAddressOfBank string
}

func (sab *CustomerBankAccountDetails) String() {
	fmt.Printf("%+v \n", sab)
}

func (cd *CustomerDetails) String() {
	fmt.Printf("%+v \n", cd)
}
