package models

// Account defines the model for accounts table
// An Account can have many transactions, AccountID is the foreign key
type Account struct {
	ID           string `gorm:"primary_key"`
	Name         string
	Balance      float64
	Transactions []Transaction `gorm:"foreignkey:AccountID"`
}

// Transaction defines the model for transactions table
type Transaction struct {
	TxID                      string
	From, To, TransactionType string
	Name                      string
	Amount                    float64
	AccountID                 string
}

// Bank structure defines the model for banks table
type Bank struct {
	Name               string `gorm:"primary_key"`
	StellarAppURL      string
	DistributorAddress string
}

// User structure defines the model for users table
type User struct {
	Name          string
	BankInfo      Bank `gorm:"foreignkey:BankName"`
	BankName      string
	BankAccountID string
	PhoneNumber   string `gorm:"primary_key"`
}
