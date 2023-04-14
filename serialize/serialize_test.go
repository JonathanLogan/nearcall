package serialize

import (
	// "fmt"
	"testing"
)

func TestEncoding(t *testing.T) {
	d, err := Format(Borsh).Serialize([]byte("TestData"))
	if err != nil {
		t.Fatalf("Serialize: %s", err)
	}
	_ = d
	// fmt.Println(string(d))
}
