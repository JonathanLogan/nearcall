package neartx

import (
	"crypto"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/near/borsh-go"
)

const (
	MaxGas         = 300_000_000_000_000
	ED25519KeyType = 0
)

const (
	CreateContractEnum = iota
	DeployContractEnum
	FunctionCallEnum
	TransferEnum
	StakeEnum
	AddKeyEnum
	DeleteKeyEnum
	DeleteAccountEnum
)

var zeroBigInt = new(big.Int)

type CreateAccount struct{}
type DeployContract struct{}
type Transfer struct{}
type Stake struct{}
type AddKey struct{}
type DeleteKey struct{}
type DeleteAccount struct{}

type FunctionCall struct {
	MethodName string
	Args       []byte
	Gas        uint64
	Deposit    big.Int
}

type Action struct {
	Enum           borsh.Enum `borsh_enum:"true"`
	CreateAccount  CreateAccount
	DeployContract DeployContract
	FunctionCall   FunctionCall
	Transfer       Transfer
	Stake          Stake
	AddKey         AddKey
	DeleteKey      DeleteKey
	DeleteAccount  DeleteAccount
}

type PublicKey struct {
	KeyType uint8
	Data    [32]byte
}

type Transaction struct {
	SignerID   string
	PublicKey  PublicKey
	Nonce      uint64
	ReceiverID string
	BlockHash  [32]byte
	Actions    []Action
}

type Signature struct {
	KeyType uint8
	Data    [64]byte
}

type SignedTransaction struct {
	Transaction Transaction
	Signature   Signature
}

type KeyPairFunc func(signer string) (keyPairFunc NamedKeyPairFunc, err error)
type NamedKeyPairFunc func() (signerName string, publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey)
type NamedPubKeyFunc func() (signerName string, publicKey ed25519.PublicKey)
type NonceFunc func(signer NamedPubKeyFunc) (nonce uint64, blockHash []byte, err error)

type NearTransaction struct {
	ETX        SignedTransaction
	privateKey NamedKeyPairFunc
	err        error
}

// NewTransaction creates a new NearTransaction sent by "signer" to "receiver". "signerKeyPair" returns the signer's keypair.
func NewTransaction(signer, receiver string, signerKeyPair KeyPairFunc) *NearTransaction {
	tx := &NearTransaction{
		ETX: SignedTransaction{
			Transaction: Transaction{
				SignerID: signer,
				PublicKey: PublicKey{
					KeyType: ED25519KeyType,
				},
				ReceiverID: receiver,
			},
			Signature: Signature{
				KeyType: ED25519KeyType,
			},
		},
	}
	kpf, err := signerKeyPair(signer)
	if err != nil {
		tx.err = err
		return tx
	}
	tx.privateKey = kpf
	_, pubKey, _ := tx.privateKey()
	copy(tx.ETX.Transaction.PublicKey.Data[:], pubKey[:])
	return tx
}

// AddFunctionCall adds a function call to "method" with "encodedArgs" as argument. Only "method" is required to be not nil.
func (tx *NearTransaction) AddFunctionCall(method string, encodedArgs []byte, gas uint64, deposit *big.Int) *NearTransaction {
	if len(method) == 0 {
		tx.err = fmt.Errorf("neartx: missing method name")
		return tx
	}
	if deposit == nil {
		deposit = zeroBigInt
	}
	if gas == 0 {
		gas = MaxGas
	}
	if encodedArgs == nil {
		encodedArgs = make([]byte, 0, 0)
	}
	tx.ETX.Transaction.Actions = append(tx.ETX.Transaction.Actions, Action{
		Enum: FunctionCallEnum,
		FunctionCall: FunctionCall{
			MethodName: method,
			Gas:        gas,
			Deposit:    *deposit,
			Args:       encodedArgs,
		},
	})
	return tx
}

func (tx *NearTransaction) Error() error {
	return tx.err
}

// Sign a NearTransaction with the keypair return from the NearTransaction() signerKeyPair.
func (tx *NearTransaction) Sign(nonceSource NonceFunc) ([]byte, error) {
	if tx.err != nil {
		return nil, tx.err
	}
	_, _, privateKey := tx.privateKey()
	if privateKey == nil {
		tx.err = fmt.Errorf("neartx: private key not available")
		return nil, tx.err
	}
	nonce, blockHash, err := nonceSource(func() (string, ed25519.PublicKey) { s, p, _ := tx.privateKey(); return s, p })
	if err != nil {
		tx.err = err
		return nil, tx.err
	}

	tx.ETX.Transaction.Nonce = nonce + 1
	copy(tx.ETX.Transaction.BlockHash[:], blockHash)
	txSer, err := borsh.Serialize(tx.ETX.Transaction)
	if err != nil {
		tx.err = err
		return nil, err
	}
	txHash := sha256.Sum256(txSer)
	sig, err := privateKey.Sign(nil, txHash[:], crypto.Hash(0))
	if err != nil {
		tx.err = err
		return nil, err
	}
	copy(tx.ETX.Signature.Data[:], sig)
	signedTx, err := borsh.Serialize(*tx)
	if err != nil {
		tx.err = err
		return nil, err
	}
	return signedTx, nil
}

// Send a NearTransaction. Signs it and sends it to rpc. nonceSource can be nil to be set to the rpc.
func (tx *NearTransaction) Send(rpc RPCURL, nonceSource NonceFunc) ([]byte, error) {
	if nonceSource == nil {
		nonceSource = rpc.GetNonce
	}
	signed, err := tx.Sign(nonceSource)
	if err != nil {
		tx.err = err
		return nil, err
	}
	return rpc.SendTransaction(signed)
}
