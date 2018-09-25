package db

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

// StellarAddressesOfBank is
type StellarAddressesOfBank struct {
	SourceSeed, IssuerSeed, DistributorSeed string
}

// CustomerBankAccountDetails is
type CustomerBankAccountDetails struct {
	Name, Balance, AccountID string
}

func decodeStellarAdressesOfBank(data []byte) (*StellarAddressesOfBank, error) {
	var bsa = &StellarAddressesOfBank{}
	err := json.Unmarshal(data, bsa)
	if err != nil {
		return nil, err
	}
	return bsa, nil
}

func decodeCustomerBankAccountDetails(data []byte) (*CustomerBankAccountDetails, error) {
	var cbad = &CustomerBankAccountDetails{}
	err := json.Unmarshal(data, cbad)
	if err != nil {
		return nil, err
	}
	return cbad, nil
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

	var customerAccountDetails *CustomerBankAccountDetails
	// Read the value back from a separate read-only transaction.
	if err := db.View(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			key := []byte(customerBankAccountID)

			customerAccountDetails, err = decodeCustomerBankAccountDetails(bucket.Get(key))
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
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
				Balance:   "10000000",
				Name:      bankName,
				AccountID: bankName,
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

	var sab *StellarAddressesOfBank
	if err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("stellarAddresses"))
			key := []byte(bankName)

			sab, err = decodeStellarAdressesOfBank(b.Get(key))
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return sab, nil
}
