package quote

import (
	"fmt"

	"github.com/edgelesssys/ego/enclave"
)

func GetQuote(message []byte) []byte {
	quote, err := enclave.GetRemoteReport([]byte(message))
	if err != nil {
		panic(err)
	}
	fmt.Println("Remote quote successfully get")

	return quote
}

// func VerifyQuote(message message.Message) attestation.Report {
// 	quote, err := enclave.VerifyRemoteReport(message.Content)
// 	if err != nil {
// 		panic(err)
// 	}

// 	publicKeyHash := sha256.Sum256(message.PublicKey)
// 	// Quote data is padded with zeros; we need to slice it to compare.
// 	if !bytes.Equal(publicKeyHash[:], quote.Data[:32]) {
// 		panic(errors.New("unequal publickey"))
// 	}

// 	selfReport, err := enclave.GetSelfReport()
// 	if err != nil {
// 		panic(err)
// 	}

// 	if !bytes.Equal(quote.SignerID, selfReport.SignerID) {
// 		panic(errors.New("invalid signer"))
// 	}
// 	if quote.Debug && !selfReport.Debug {
// 		panic(errors.New("other party is a debug enclave"))
// 	}

// 	return quote
// }
