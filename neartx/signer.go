package neartx

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"github.com/mr-tron/base58/base58"
	"os"
	"path"
)

func GetSigner(name, network string) (*Signer, error) {
	hd, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fn := path.Join(hd, ".near-credentials", network, name) + ".json"
	content, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	signer := new(Signer)
	if err := json.Unmarshal(content, signer); err != nil {
		return nil, err
	}
	if err := signer.parse(); err != nil {
		return nil, err
	}
	return signer, nil
}

type Signer struct { // ToDo: Generalize for other signature algorithms.
	AccountID  string `json:"account_id"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

func (signer *Signer) parse() error {
	const preamble = "ed25519:"
	if signer.publicKey != nil && signer.privateKey != nil && signer.AccountID != "" {
		return nil
	}
	if signer.PrivateKey[0:len(preamble)] == preamble {
		pk, err := base58.Decode(signer.PrivateKey[len(preamble):])
		if err != nil {
			return err
		}
		signer.privateKey = ed25519.PrivateKey(pk)
	} else {
		return fmt.Errorf("neartx: malformed private key for <%s>", signer.AccountID)
	}
	if signer.PublicKey[0:len(preamble)] == preamble {
		pk, err := base58.Decode(signer.PublicKey[len(preamble):])
		if err != nil {
			return err
		}
		signer.publicKey = ed25519.PublicKey(pk)
	} else {
		signer.publicKey = ed25519.PublicKey(signer.privateKey[32:])
	}
	return nil
}

func (signer *Signer) KeyPair(signerReq string) (keypairFunc NamedKeyPairFunc, err error) {
	if err := signer.parse(); err != nil {
		return nil, err
	}
	return func() (string, ed25519.PublicKey, ed25519.PrivateKey) {
		return signer.AccountID, signer.publicKey, signer.privateKey
	}, nil
}

type Network string

func (network Network) FileSigner(signer string) (keypairFunc NamedKeyPairFunc, err error) {
	s, err := GetSigner(signer, string(network))
	if err != nil {
		return nil, err
	}
	return s.KeyPair(signer)
}
