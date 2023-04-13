package main

import (
	"flag"
	"fmt"
	"github.com/JonathanLogan/nearcall/neartx"
	json "github.com/buger/jsonparser"
	"os"
	"strings"
)

const (
	networkPlaceholder = "%"
)

var (
	Network  string = "mainnet"
	RPC      string = "https://rpc." + networkPlaceholder + ".near.org/"
	Signer   string
	Receiver string
	Method   string
	ArgFile  string
	Encoding string = "binary"
	Verbose  bool
)

func init() {
	flag.StringVar(&Network, "network", Network, "mainnet/testnet NEAR network")
	flag.StringVar(&RPC, "rpc", RPC, "URL of RPC node. \""+networkPlaceholder+"\" will be replaced with network")
	flag.StringVar(&Signer, "signer", "", "Signer/Sender of transaction")
	flag.StringVar(&Receiver, "receiver", "", "Receiver of transaction")
	flag.StringVar(&Method, "method", "", "Name of method to call")
	flag.StringVar(&ArgFile, "arg", "", "File containing the method argument")
	flag.StringVar(&Encoding, "encoding", Encoding, "Encoding for arg: binary, hex, xHex, 0xHex, base64, base58, borsh")
	flag.BoolVar(&Verbose, "v", false, "Enable verbose transaction output")
}

func validateFlags() {
	var ok bool = true
	if Network == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Missing network argument\n")
		ok = false
	}
	if RPC == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Missing rpc argument\n")
		ok = false
	}
	if Signer == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Missing signer argument\n")
		ok = false
	}
	if Receiver == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Missing receiver argument\n")
		ok = false
	}
	if Method == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Missing method argument\n")
		ok = false
	}
	if ArgFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Missing arg argument\n")
		ok = false
	}
	if Encoding == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Missing encoding argument\n")
		ok = false
	}
	if !ok {
		os.Exit(1)
	}
}

func EncodeFile(filename, encoding string) ([]byte, error) {
	d, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return neartx.StrFormat(encoding).Serialize(d)
}

func extractError(d []byte) []byte {
	var errorData []byte
	json.EachKey(d, func(idx int, value []byte, vt json.ValueType, err error) {
		switch idx {
		case 0:
			errorData = make([]byte, len(value))
			copy(errorData, value)
		}
	}, []string{"result", "receipts_outcome", "[0]", "outcome", "status", "Failure"})
	return errorData
}

func main() {
	flag.Parse()
	validateFlags()
	trueRPC := strings.ReplaceAll(RPC, networkPlaceholder, Network)

	encoded, err := EncodeFile(ArgFile, Encoding)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading and encoding %s: %s\n", ArgFile, err)
		os.Exit(1)
	}

	tx := neartx.NewTransaction(Signer, Receiver, neartx.KeyPairFunc(neartx.Network(Network).FileSigner)).AddFunctionCall(Method, encoded, 0, nil)
	ret, err := tx.Send(neartx.RPCURL(trueRPC), nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error sending transaction: %s", err)
		os.Exit(1)
	}
	if eData := extractError(ret); eData != nil {
		_, _ = fmt.Fprintln(os.Stderr, string(eData))
		os.Exit(2)
	}
	if Verbose {
		_, _ = fmt.Fprintln(os.Stdout, string(ret))
	}
	os.Exit(0)
}
