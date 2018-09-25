package account

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
)

// Account is a custom structure type that aggregates 2 string fields that SHOULD be exportable
type Account struct{ Seed, Address string }

// MakePair returns
func MakePair() *keypair.Full {
	pair, err := keypair.Random()
	if err != nil {
		log.Fatal(err)
	}

	// log.Println(pair.Seed())
	// log.Println(pair.Address())

	// file, err := os.Create("./testAccounts.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// file.WriteString(pair.Address+'')

	return pair
}

// CreateTestAccount returns a test accont key pair and account details
func CreateTestAccount() (*keypair.Full, error) {
	// pair is the pair that was generated from previous example, or create a pair based on
	// existing keys.
	pair := MakePair()
	address := pair.Address()
	_, err := http.Get("https://friendbot.stellar.org/?addr=" + address)
	if err != nil {
		return nil, err
	}
	return pair, nil
	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(string(body))
	// return pair, body
	/*
		{
		  "_links": {
		    "transaction": {
		      "href": "https://horizon-testnet.stellar.org/transactions/723b27f4de27f3343370d9035a23af63f6dbdb088294ee4bba2f348d03f7e31e"
		    }
		  },
		  "hash": "723b27f4de27f3343370d9035a23af63f6dbdb088294ee4bba2f348d03f7e31e",
		  "ledger": 10629716,
		  "envelope_xdr": "AAAAABB90WssODNIgi6BHveqzxTRmIpvAFRyVNM+Hm2GVuCcAAAAZABiwhcAA0QFAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAkLpH4c8xeifbL+ywyaaUXHptwPTqKIBMyT5H/91L+IoAAAAXSHboAAAAAAAAAAABhlbgnAAAAECwP1Zv0N+hRYHhHdyf1XaCbYPWhIdstOvv3YiLdRx/ShZcfOOus6iMTQ1KhiOiOwTzgjP8+OhtxxUyy5cUgWIL",
		  "result_xdr": "AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAA=",
		  "result_meta_xdr": "AAAAAQAAAAIAAAADAKIyVAAAAAAAAAAAEH3Rayw4M0iCLoEe96rPFNGYim8AVHJU0z4ebYZW4JwAFDmtgrgP/wBiwhcAA0QEAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAKIyVAAAAAAAAAAAEH3Rayw4M0iCLoEe96rPFNGYim8AVHJU0z4ebYZW4JwAFDmtgrgP/wBiwhcAA0QFAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAojJUAAAAAAAAAACQukfhzzF6J9sv7LDJppRcem3A9OoogEzJPkf/3Uv4igAAABdIdugAAKIyVAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAMAojJUAAAAAAAAAAAQfdFrLDgzSIIugR73qs8U0ZiKbwBUclTTPh5thlbgnAAUOa2CuA//AGLCFwADRAUAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAojJUAAAAAAAAAAAQfdFrLDgzSIIugR73qs8U0ZiKbwBUclTTPh5thlbgnAAUOZY6QSf/AGLCFwADRAUAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA=="
		}
	*/
	/*
		{
		  "_links": {
		    "transaction": {
		      "href": "https://horizon-testnet.stellar.org/transactions/dc1a77edd6c7b9ba931fc3a120cfa8ef3482ecb8a6cf4f87b00d0b0e1ad14b67"
		    }
		  },
		  "hash": "dc1a77edd6c7b9ba931fc3a120cfa8ef3482ecb8a6cf4f87b00d0b0e1ad14b67",
		  "ledger": 10633793,
		  "envelope_xdr": "AAAAABB90WssODNIgi6BHveqzxTRmIpvAFRyVNM+Hm2GVuCcAAAAZABiwhcAA0YCAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAQXbhB4yCPAwbo1Qr1ptCXKoY5WS/GZBG6IcuDlfQwWgAAAAXSHboAAAAAAAAAAABhlbgnAAAAEBaQYzjgyWYnCMSNdxzM8Pn8PjXpF60EIM2SrET4uCgQp2bs0LgY8hsSaJnxzBD6/B5SLA1gbzOBfRDnv7S+Q0M",
		  "result_xdr": "AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAA=",
		  "result_meta_xdr": "AAAAAQAAAAIAAAADAKJCQQAAAAAAAAAAEH3Rayw4M0iCLoEe96rPFNGYim8AVHJU0z4ebYZW4JwAFAzuPjFpKwBiwhcAA0YBAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAKJCQQAAAAAAAAAAEH3Rayw4M0iCLoEe96rPFNGYim8AVHJU0z4ebYZW4JwAFAzuPjFpKwBiwhcAA0YCAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAwAAAAAAokJBAAAAAAAAAABBduEHjII8DBujVCvWm0JcqhjlZL8ZkEbohy4OV9DBaAAAABdIdugAAKJCQQAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAMAokJBAAAAAAAAAAAQfdFrLDgzSIIugR73qs8U0ZiKbwBUclTTPh5thlbgnAAUDO4+MWkrAGLCFwADRgIAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAokJBAAAAAAAAAAAQfdFrLDgzSIIugR73qs8U0ZiKbwBUclTTPh5thlbgnAAUDNb1uoErAGLCFwADRgIAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA=="
		}
	*/
}

// GetCreatedAccounts returns created accounts
func GetCreatedAccounts() struct{ PersonA, PersonB Account } {
	accounts := struct{ PersonA, PersonB Account }{
		Account{
			Seed:    "SCJLOWB3IVYS4EEGUTITFIHL2KDSRP5YNYKGSD7OCUQOBUBXV4Q5ZIVZ", // has to an exported field so has to start with capital letter
			Address: "GCILUR7BZ4YXUJ63F7WLBSNGSROHU3OA6TVCRACMZE7EP765JP4IUOM4",
		},
		Account{
			Seed:    "SAT6YI3GYOOU4UX2WAWNJSXQ3YBA4GJBJCUTR4MSAGNFXFUB3MZIAAZT",
			Address: "GBAXNYIHRSBDYDA3UNKCXVU3IJOKUGHFMS7RTECG5CDS4DSX2DAWQGNH",
		},
	}
	return accounts
}

// PrintAccountDetails prints account details
func PrintAccountDetails(address string) {
	fmt.Printf("fetching account details for account address: %s .... \n\n", address)
	account, err := horizon.DefaultTestNetClient.LoadAccount(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Balances for account:", address)

	for _, balance := range account.Balances {
		log.Println(balance)
	}
	fmt.Println()
}
