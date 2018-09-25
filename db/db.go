package db

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
)

// StellarAddressesOfBank is

type data interface {
	String()
}

//StellarAddressesOfBank implements data, contains
type StellarAddressesOfBank struct {
	SourceSeed, IssuerSeed, DistributorSeed string
}

//TransactionDetails contains
type TransactionDetails struct {
	From, To, Amount, TransactionID string
}

//CustomerBankAccountDetails implements data, contains
type CustomerBankAccountDetails struct {
	Name         string
	Balance      float64
	Transactions []TransactionDetails
}

//CustomerDetails implements data, contains
type CustomerDetails struct {
	CustomerName, BankName, BankAccountID string
}

func (sab *StellarAddressesOfBank) String() {
	fmt.Printf("%+v \n", sab)
}

func (sab *CustomerBankAccountDetails) String() {
	fmt.Printf("%+v \n", sab)
}

func (sab *CustomerDetails) String() {
	fmt.Printf("%+v \n", sab)
}

//decodeByteSlice returns
func decodeByteSlice(dBdata []byte, target data) (data, error) {
	err := json.Unmarshal(dBdata, target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

// UpdateCustomerBankAccountBalence returns
func UpdateCustomerBankAccountBalence(bankName, customerBankAccountID string, newBalance float64) error {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	bucketName := "AccountDetails"

	// Execute several commands within a read-write transaction.
	if err := db.Update(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			key := []byte(customerBankAccountID)

			dataVal, err := decodeByteSlice(bucket.Get(key), &CustomerBankAccountDetails{})
			if err != nil {
				return err
			}
			accountDetails, ok := dataVal.(*CustomerBankAccountDetails)
			if !ok {
				return errors.New("Could not update Customer Bank Account Details : Type assertion failed")
			}

			accountDetails.Balance = newBalance

			encoded, err := json.Marshal(accountDetails)
			if err != nil {
				return err
			}

			if err := bucket.Put(key, encoded); err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

// WriteCustomerBankAccountDetails returns
func WriteCustomerBankAccountDetails(bankName string, customerBankAccountID string, customerBankAccountDetails *CustomerBankAccountDetails) error {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	bucketName := "AccountDetails"

	// Execute several commands within a read-write transaction.
	if err := db.Update(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			key := []byte(customerBankAccountID)

			encoded, err := json.Marshal(customerBankAccountDetails)
			if err != nil {
				return err
			}

			if err := bucket.Put(key, encoded); err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

// ReadCustomerBankAccountDetails returns
func ReadCustomerBankAccountDetails(bankName, customerBankAccountID string) (*CustomerBankAccountDetails, error) {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	bucketName := "AccountDetails"

	var accountDetails data
	// Read the value back from a separate read-only transaction.
	if err := db.View(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			key := []byte(customerBankAccountID)

			accountDetails, err = decodeByteSlice(bucket.Get(key), &CustomerBankAccountDetails{})
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}
	customerAccountDetails, ok := accountDetails.(*CustomerBankAccountDetails)
	if !ok {
		return nil, errors.New("Could not read Customer Bank Account Details : Type assertion failed")
	}
	return customerAccountDetails, nil
}

// CreateDBForBank returns
func CreateDBForBank(bankName string) error {
	db, err := bolt.Open(bankName+".db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			b, err := tx.CreateBucket([]byte("AccountDetails"))
			if err != nil {
				return err
			}

			encoded, err := json.Marshal(&CustomerBankAccountDetails{
				Balance: 10000000.0,
				Name:    bankName,
			})
			if err != nil {
				return err
			}

			if err := b.Put([]byte(bankName), encoded); err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

// CreateStellarAddressesOfBankDB returns
func CreateStellarAddressesOfBankDB() error {
	db, err := bolt.Open("BankStellarAddresses.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("StellarAddresses"))
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

// WriteStellarAddressesForBank returns
func WriteStellarAddressesForBank(bankName string, addresses *StellarAddressesOfBank) error {
	db, err := bolt.Open("BankStellarAddresses.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("StellarAddresses"))
			key := []byte(bankName)

			encoded, err := json.Marshal(addresses)
			if err != nil {
				return err
			}

			err = b.Put(key, encoded)
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

// ReadStellarAddressesOfBank returns
func ReadStellarAddressesOfBank(bankName string) (*StellarAddressesOfBank, error) {
	db, err := bolt.Open("BankStellarAddresses.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var dataVal data
	if err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("StellarAddresses"))
			key := []byte(bankName)

			dataVal, err = decodeByteSlice(b.Get(key), &StellarAddressesOfBank{})
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	stellarAddress, ok := dataVal.(*StellarAddressesOfBank)
	if !ok {
		return nil, errors.New("Could not read Stellar Bank Address : Type assertion failed")
	}
	return stellarAddress, nil
}

// CreateCommonCustomersDB returns
func CreateCommonCustomersDB() error {
	db, err := bolt.Open("CommonCustomers.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("CustomerDetails"))
			if err != nil {
				return err
			}
			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

//ReadCustomerDetailsFromCommonCustomersDB returns
func ReadCustomerDetailsFromCommonCustomersDB(phoneNumber string) (*CustomerDetails, error) {
	db, err := bolt.Open("CommonCustomers.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var dataVal data
	if err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("CustomerDetails"))
			key := []byte(phoneNumber)

			dataVal, err = decodeByteSlice(b.Get(key), &StellarAddressesOfBank{})
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	cd, ok := dataVal.(*CustomerDetails)
	if !ok {
		return nil, errors.New("Could not read Customer Details: Type assertion failed")
	}

	return cd, nil
}

// WriteCustomerDetailsToCommonCustomersDB returns
func WriteCustomerDetailsToCommonCustomersDB(phoneNumber string, customerDetails *CustomerDetails) error {
	db, err := bolt.Open("CommonCustomers.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("CustomerDetails"))
			key := []byte(phoneNumber)

			encoded, err := json.Marshal(customerDetails)
			if err != nil {
				return err
			}

			err = b.Put(key, encoded)
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}
