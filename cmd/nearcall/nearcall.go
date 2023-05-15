package main

import (
	"flag"
	"fmt"
	"github.com/JonathanLogan/nearcall/neartx"
	"github.com/JonathanLogan/nearcall/serialize"
	json "github.com/buger/jsonparser"
	"math/big"
	"os"
	"strings"
)

const (
	networkPlaceholder = "%"
)

var (
	Network        string = "mainnet"
	RPC            string = "https://rpc." + networkPlaceholder + ".near.org/"
	Signer         string
	Receiver       string
	Method         string
	ArgFile        string
	Arg            string
	OutputEncoding string = "binary"
	InputEncoding  string = "binary"
	Verbose        bool
	Deposit        string
)

func init() {
	flag.StringVar(&Network, "network", Network, "mainnet/testnet NEAR network")
	flag.StringVar(&RPC, "rpc", RPC, "URL of RPC node. \""+networkPlaceholder+"\" will be replaced with network")
	flag.StringVar(&Signer, "signer", "", "Signer/Sender of transaction")
	flag.StringVar(&Receiver, "receiver", "", "Receiver of transaction")
	flag.StringVar(&Method, "method", "", "Name of method to call")
	flag.StringVar(&ArgFile, "argfile", "", "File containing the method argument")
	flag.StringVar(&Arg, "arg", "", "Method argument")
	flag.StringVar(&OutputEncoding, "out", OutputEncoding, "Output encoding for arg: binary, hex, xHex, 0xHex, base64, base58, borsh")
	flag.StringVar(&InputEncoding, "in", InputEncoding, "Input encoding for arg: binary, hex, xHex, 0xHex, base64, base58, borsh")
	flag.StringVar(&Deposit, "deposit", "", "Attach yoctoNear")
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
	if ArgFile == "" && Arg == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Either arg or argfile need to be given\n")
		ok = false
	}
	if ArgFile != "" && Arg != "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Only one of arg or argfile permitted\n")
		ok = false
	}
	if OutputEncoding == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Missing output encoding argument\n")
		ok = false
	}
	if InputEncoding == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Missing input encoding argument\n")
		ok = false
	}
	if Deposit != "" {
		if _, ok := new(big.Int).SetString(Deposit, 10); !ok {
			_, _ = fmt.Fprintf(os.Stderr, "Error: Deposit cannot be parsed as decimal integer\n")
		}
	}
	if !ok {
		os.Exit(1)
	}
}

func GetFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
	// if err != nil {
	// return nil, err
	// }
	// return neartx.StrFormat(encoding).Serialize(d)
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
	var encoded []byte
	var deposit *big.Int
	flag.Parse()
	validateFlags()
	deposit, _ = new(big.Int).SetString(Deposit, 10)
	trueRPC := strings.ReplaceAll(RPC, networkPlaceholder, Network)

	{
		var err error
		var input []byte
		if ArgFile != "" {
			if input, err = GetFile(ArgFile); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
				os.Exit(1)
			}
			// encoded, err = EncodeFile(ArgFile, OutputEncoding)
		} else if Arg != "" {
			input = []byte(Arg)
		}
		input, err = serialize.StrFormat(InputEncoding).Deserialize(input)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error decoding input: %s\n", err)
			os.Exit(1)
		}
		encoded, err = serialize.StrFormat(OutputEncoding).Serialize(input)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error encoding %s: %s\n", ArgFile, err)
			os.Exit(1)
		}
	}

	tx := neartx.NewTransaction(Signer, Receiver, neartx.KeyPairFunc(neartx.Network(Network).FileSigner)).AddFunctionCall(Method, encoded, 0, deposit)
	ret, err := tx.Send(neartx.RPCURL(trueRPC), nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error sending transaction: %s\n", err)
		os.Exit(2)
	}
	if eData := extractError(ret); eData != nil {
		_, _ = fmt.Fprintln(os.Stderr, string(eData))
		os.Exit(3)
	}
	if Verbose {
		_, _ = fmt.Fprintln(os.Stdout, string(ret))
	}
	os.Exit(0)
}
