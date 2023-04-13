package neartx

import (
	// "fmt"
	"testing"
)

func TestNearTx(t *testing.T) {
	t.Skip()
	tx := NewTransaction("valconaur.near", "aurora", KeyPairFunc(Network("mainnet").FileSigner)).AddFunctionCall("submit", []byte("testing"), 0, nil)
	if err := tx.Error(); err != nil {
		t.Fatalf("NewTX: %s", err)
	}
	ret, err := tx.Send(RPCURL("https://rpc.mainnet.near.org/"), nil)
	if err != nil {
		t.Fatalf("Send: %s", err)
	}
	if err := tx.Error(); err != nil {
		t.Fatalf("Send2: %s", err)
	}
	_ = ret
	// fmt.Println(string(ret))
}
