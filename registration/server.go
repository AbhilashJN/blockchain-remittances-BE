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

func getUserInfo(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "r.ParseForm() failed: %v", err)
		}

		phoneNumberQueryKey := "PhoneNumber"

		if _, ok := r.Form[phoneNumberQueryKey]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s parameter not found in the query string", phoneNumberQueryKey)
			return
		}

		phoneNumber := r.FormValue(phoneNumberQueryKey)

		var user models.User

		if err := db.Where("phone_number = ?", phoneNumber).Preload("BankInfo").First(&user).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "No user with %s as phone number found", phoneNumber)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `db.Where("phone_number = ?", phoneNumber).Preload("BankInfo").First(&user).Error failed: %v`, err)
			return
		}

		// fmt.Printf("%+v\n", user)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			fmt.Fprintf(w, "json.NewEncoder(w).Encode(user) failed:\n errorL %v", err)
			return
		}
	}
}

func registerNewUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		// fmt.Printf("%+v\n", r.PostForm)
		bankNameQuerykey, bankAccIDQuerykey, phoneNumberQuerykey := "BankName", "BankAccountID", "PhoneNumber"

		for _, fieldName := range []string{bankNameQuerykey, bankAccIDQuerykey, phoneNumberQuerykey} {
			if _, ok := r.PostForm[fieldName]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "%s field not found in the request's body", fieldName)
				return
			}
		}

		bankName, bankAccountID, phoneNumber := r.PostFormValue(bankNameQuerykey), r.PostFormValue(bankAccIDQuerykey), r.PostFormValue(phoneNumberQuerykey)

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
			fmt.Fprintf(w, `db.Model(&user).Related(&userBankInfo, "BankInfo").Error failed:\n %v`, err)
			return
		}

		user.BankInfo = userBankInfo

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "json.NewEncoder(w).Encode(user) failed:\n %v", err)
			return
		}
	}
}

func getBanksList(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method == "GET" {
		var bankNameList []string
		rows, err := db.Table("banks").Select("name").Rows()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `db.Table("banks").Select("name").Rows() failed:\n %v`, err)
			return
		}
		for rows.Next() {
			var bankName string
			rows.Scan(&bankName)
			bankNameList = append(bankNameList, bankName)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(bankNameList); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "json.NewEncoder(w).Encode(user) failed:\n %v", err)
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
	AppServer ServerConfig
	Database  DatabaseConfig
}

var config Config

var seedBanks = []models.Bank{
	models.Bank{Name: "ALPHA", StellarAppURL: "localhost:7070", DistributorAddress: "GALRIXDRLNMOAUOQLNEFC5TJ5Q4M5WW7WLVOIJRJAGFNJBQHJMJMS635", NativeCurrency: "INR"},
	models.Bank{Name: "BETA", StellarAppURL: "localhost:6060", DistributorAddress: "GAJPTAGSHHH5ZSRCEZIQTYHXZZSOUTKAV7342ONN4DQVGY7PR2TDE2BP", NativeCurrency: "USD"},
}

var seedUsers = []models.User{
	{PhoneNumber: "9976543210", Name: "Sreekar", BankName: "ALPHA", BankAccountID: "123ABC"},
	// {PhoneNumber: "9876543210", Name: "Abhilash", BankName: "SBI", BankAccountID: "456DEF"},
	{PhoneNumber: "8976543210", Name: "Milan", BankName: "BETA", BankAccountID: "789GHI"},
	// {PhoneNumber: "8876543210", Name: "Sandeep", BankName: "JP MORGAN", BankAccountID: "321KLM"},
}

func seedTables(db *gorm.DB) {
	for _, bank := range seedBanks {
		if err := db.Create(&bank).Error; err != nil {
			fmt.Println(err)
		}
	}
	for _, user := range seedUsers {
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

	serverAddress := fmt.Sprintf("localhost:%s", config.AppServer.Port)
	print(serverAddress)
	http.HandleFunc("/registerNewUser", makeHandler(registerNewUser, db))
	http.HandleFunc("/getUserInfo", makeHandler(getUserInfo, db))
	http.HandleFunc("/getBanksList", makeHandler(getBanksList, db))
	fmt.Println("\n\nRegistartion server is starting...")
	err = http.ListenAndServe(serverAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
