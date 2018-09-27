package utils

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/stellar/go/xdr"
)

// DecodeTransactionEnvelope returns
func DecodeTransactionEnvelope(data string) (xdr.TransactionEnvelope, error) {

	rawr := strings.NewReader(data)
	b64r := base64.NewDecoder(base64.StdEncoding, rawr)

	var tx xdr.TransactionEnvelope
	bytesRead, err := xdr.Unmarshal(b64r, &tx)

	if err != nil {
		return tx, err
	}

	fmt.Printf("Successful decoding of transaction envelope. Read %d bytes\n", bytesRead)
	return tx, nil
}
