package neartx

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mr-tron/base58/base58"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	maxBodyLength = 4 * (2 << 19) // 4MB
)

type RPCURL string

func (rpcUrl RPCURL) GetNonce(f NamedPubKeyFunc) (nonce uint64, blackHash []byte, err error) {
	accountID, publicKey := f()
	if publicKey == nil {
		return 0, nil, fmt.Errorf("neartx: public key not available for <%s>", accountID)
	}
	return GetNonce(string(rpcUrl), accountID, publicKey)
}

func (rpcUrl RPCURL) SendTransaction(tx []byte) ([]byte, error) {
	return SendTx(string(rpcUrl), tx)
}

func createRequest(rpcUrl string, r io.Reader) *http.Request {
	req, err := http.NewRequest(http.MethodPost, rpcUrl, r)
	if err != nil {
		panic(err.Error())
	}
	req.Header.Add("Accept", "application/json, */*;q=0.5")
	req.Header.Add("Content-Type", "application/json")
	return req
}

func httpRequestViewAccessKey(rpcUrl, accountId string, pubKey ed25519.PublicKey) *http.Request {
	request := `{"id":"dontcare","jsonrpc":"2.0","method":"query","params":{"finality":"final","request_type":"view_access_key","account_id":"%s","public_key": "ed25519:%s"}}`
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintf(buf, request, accountId, base58.Encode(pubKey))
	return createRequest(rpcUrl, buf)
}

func GetNonce(rpcUrl, accountId string, pubKey ed25519.PublicKey) (uint64, []byte, error) {
	type ViewAccessResponse struct {
		Result struct {
			BlockHash   string  `json:"block_hash"`
			BlockHeight uint64  `json:"block_height"`
			Nonce       *uint64 `json:"nonce"`
			Error       *string `json:"error"`
		} `json:"result"`
	}
	var ret uint64
	resp, err := rpcCall(httpRequestViewAccessKey(rpcUrl, accountId, pubKey))
	if err != nil {
		return ret, nil, err
	}
	defer resp.Body.Close()
	v := new(ViewAccessResponse)
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(v); err != nil {
		return ret, nil, err
	}
	if v.Result.Error != nil {
		return ret, nil, errors.New(*v.Result.Error)
	}
	if v.Result.Nonce != nil {
		ret = *v.Result.Nonce
	}
	blockHash, err := base58.Decode(v.Result.BlockHash)
	if err != nil {
		return ret, nil, err
	}
	return ret, blockHash, nil
}

func httpSendTx(rpcUrl string, transaction []byte) *http.Request {
	request := `{"id":"dontcare","jsonrpc":"2.0","method":"broadcast_tx_commit","params":["%s"]}`
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintf(buf, request, base64.StdEncoding.EncodeToString(transaction))
	return createRequest(rpcUrl, buf)
}

type RPCError struct {
	RPCError *struct {
		Name  string `json:"name"`
		Cause struct {
			Info map[string]interface{} `json:"info"`
			Name string                 `json:"name"`
		} `json:"cause"`
		Code    int64  `json:"code"`
		Message string `json:"message"`
		Data    struct {
			TxExecutionError map[string]interface{}
		} `json:"data"`
	} `json:"error"`
}

func (err *RPCError) Error() string {
	return fmt.Sprintf("rpc: %s %s", err.RPCError.Name, err.RPCError.Cause.Name)
}

func rpcCall(r *http.Request) (*http.Response, error) {
	client := *http.DefaultClient
	client.Timeout = time.Second * 10
	return client.Do(r)
}

func SendTx(rpcUrl string, transaction []byte) ([]byte, error) {
	var rpcError RPCError
	resp, err := rpcCall(httpSendTx(rpcUrl, transaction))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	d, err := ioutil.ReadAll(io.LimitReader(resp.Body, maxBodyLength))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(d, &rpcError); err != nil {
		return nil, err
	}
	if rpcError.RPCError != nil {
		return nil, &rpcError
	}
	return d, nil
}
