package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/AbhilashJN/blockchain-remittances-BE/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

func registerNewUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		fmt.Printf("%+v\n", r.PostForm)

		for _, fieldName := range []string{"BankName", "BankAccountID", "PhoneNumber"} {
			_, ok := r.PostForm[fieldName]
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "%s field not found in the request's body", fieldName)
				return
			}
		}

		bankName, bankAccountID, phoneNumber := r.PostFormValue("BankName"), r.PostFormValue("BankAccountID"), r.PostFormValue("PhoneNumber")

		var bankInfo models.Bank
		if err := db.Where("Name = ?", bankName).First(&bankInfo).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, fmt.Sprintf("%q is an invalid bank name ", bankName))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "db.Where(&models.Bank{Name: bankName}).First(&bankInfo) failed:\n %v", err)
			return
		}

		resp, err := http.Get("http://" + bankInfo.StellarAppURL + "/accountDetails?BankAccountID=" + bankAccountID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `http.Get(bankInfo.StellarAppURL + "/accountDetails?BankAccountID=" + bankAccountID) failed:\n %v`, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "ioutil.ReadAll(resp.Body) failed:\n %v", err)
				return
			}
			w.WriteHeader(resp.StatusCode)
			fmt.Fprintf(w, "Non 2xx response from the bank, resp.body: %s", body)
			return
		}

		var accountDetails models.Account

		if err := json.NewDecoder(resp.Body).Decode(&accountDetails); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "json.NewDecoder(resp.Body).Decode(&accountDetails) failed:\n %v", err)
			return
		}

		var user = models.User{Name: accountDetails.Name, BankName: bankName, BankAccountID: bankAccountID, PhoneNumber: phoneNumber}

		if err := db.Create(&user).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "db.Create(&user).Error failed:\n %v", err)
			return
		}

		// Verifies the association of users table with the banks table
		var userBankInfo models.Bank
		if err := db.Model(&user).Related(&userBankInfo, "BankInfo").Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `db.Where("phone_number = ?", phoneNumber).First(&userDet) failed:\n %v`, err)
			return
		}

		user.BankInfo = userBankInfo

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "json.NewEncoder(w).Encode(userDet) failed:\n %v", err)
			return
		}
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *gorm.DB), db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db)
	}
}

// ServerConfig is config for app server
type ServerConfig struct {
	Port string
}

// DatabaseConfig is config for database
type DatabaseConfig struct {
	Host, Port, User, Password, DbName string
}

// Config is config for the entire app
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

var config Config

var banks = []models.Bank{
	models.Bank{Name: "SBI", StellarAppURL: "localhost:7070", DistributorAddress: "GC2JUDOWWCREXJPNGCL4IFBF6C6EVFVEHBSJWQT26T6A63TWIIOYQZQH"},
	models.Bank{Name: "JP MORGAN", StellarAppURL: "localhost:6060", DistributorAddress: "GCG2S7CUX4VWXNW5LL3V7CGD36ZBV6TED43LN4B772M5JQ7Z7I43SEOT"},
}

var users = []models.User{
	{PhoneNumber: "9976543210", Name: "Sreekar", BankName: "SBI", BankAccountID: "123ABC"},
	// {PhoneNumber: "9876543210", Name: "Abhilash", BankName: "SBI", BankAccountID: "456DEF"},
	{PhoneNumber: "8976543210", Name: "Milan", BankName: "JP MORGAN", BankAccountID: "789GHI"},
	// {PhoneNumber: "8876543210", Name: "Sandeep", BankName: "JP MORGAN", BankAccountID: "321KLM"},
}

func seedTables(db *gorm.DB) {
	for _, bank := range banks {
		if err := db.Create(&bank).Error; err != nil {
			fmt.Println(err)
		}
	}
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			fmt.Println(err)
		}
	}
}

func readConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // ensures that Viper will read from environment variables as well.

	// Searches for config file in given paths and read it
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Confirm which config file is used
	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

func init() {
	readConfig()
	// fmt.Printf("%+v\n", config)

}

func main() {
	dbConfig := config.Database
	dbConnectionParams := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.DbName, dbConfig.Password)

	db, err := gorm.Open("postgres", dbConnectionParams)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	db.DropTableIfExists(&models.User{}, &models.Bank{})
	db.AutoMigrate(&models.User{}, &models.Bank{})
	db.Model(&models.User{}).AddForeignKey("bank_name", "banks(name)", "CASCADE", "CASCADE")

	seedTables(db)

	serverAddress := fmt.Sprintf("localhost:%s", config.Server.Port)
	print(serverAddress)
	http.HandleFunc("/registerNewUser", makeHandler(registerNewUser, db))

	fmt.Println("\n\nRegistartion server is starting...")
	err = http.ListenAndServe(serverAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
