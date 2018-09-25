package db

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
)

// StellarAddressesOfBank is

type data interface {
	String() string
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
	Name, Balance string
	Transactions  []TransactionDetails
}

//CustomerDetails implements data, contains
type CustomerDetails struct {
	CustomerName, BankName, BankAccountID string
}

func (sab *StellarAddressesOfBank) String() string {
	return fmt.Sprintf("%+v \n", sab)
}

func (sab *CustomerBankAccountDetails) String() string {
	return fmt.Sprintf("%+v \n", sab)
}

func (sab *CustomerDetails) String() string {
	return fmt.Sprintf("%+v \n", sab)
}

//decodeByteSlice returns
func decodeByteSlice(dBdata []byte, target data) (data, error) {
	err := json.Unmarshal(dBdata, target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

// UpdateAccountBalence returns
// func UpdateAccountBalence(bankName, customerBankAccountID, newBalance string) error {
// 	// Open the database.
// 	db, err := bolt.Open(bankName+".db", 0666, nil)
// 	if err != nil {
// 		return err
// 	}
// 	defer db.Close()
// 	bucketName := bankName + "-balances"

// 	// Execute several commands within a read-write transaction.
// 	if err := db.Update(
// 		func(tx *bolt.Tx) error {
// 			bucket := tx.Bucket([]byte(bucketName))
// 			key := []byte(customerBankAccountID)

// 			cbad, err := decodeCustomerBankAccountDetails(bucket.Get(key))
// 			if err != nil {
// 				return err
// 			}

// 			cbad.CustomerBalance = newBalance

// 			encoded, err := json.Marshal(cbad)
// 			if err != nil {
// 				return err
// 			}

// 			if err := bucket.Put(key, encoded); err != nil {
// 				return err
// 			}

// 			return nil
// 		},
// 	); err != nil {
// 		return err
// 	}

// 	return nil
// }

// GetAccountDetails returns
func GetAccountDetails(bankName, customerBankAccountID string) (*CustomerBankAccountDetails, error) {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	bucketName := bankName + "-balances"

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
			b, err := tx.CreateBucket([]byte(bankName + "-balances"))
			if err != nil {
				return err
			}

			encoded, err := json.Marshal(&CustomerBankAccountDetails{
				Balance: "10000000",
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

// func parseCLIFlags() *string {
// 	dbName := flag.String("dbName", "default", "name of db")
// 	flag.Parse()
// 	return dbName
// }

// StoreStellarAddressesOfBank returns
func StoreStellarAddressesOfBank(bankName string, adresses *StellarAddressesOfBank) error {
	db, err := bolt.Open("BankStellarAddresses.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("stellarAddresses"))
			if err != nil {
				return err
			}

			encoded, err := json.Marshal(*adresses)
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

// RetreiveStellarAddressesOfBank returns
func RetreiveStellarAddressesOfBank(bankName string) (*StellarAddressesOfBank, error) {
	db, err := bolt.Open("BankStellarAddresses.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var dataVal data
	if err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("stellarAddresses"))
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

//CreateCentralCustomerDB returns
func CreateCentralCustomerDB() error {
	db, err := bolt.Open("CentralCustomerDB"+".db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Update(
		func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("CustomerDetails"))
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

//ReadCustomerDetails returns
func ReadCustomerDetails(phoneNumber string) (*CustomerDetails, error) {
	db, err := bolt.Open("CentralCustomerDB.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var dataVal data
	if err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("stellarAddresses"))
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
