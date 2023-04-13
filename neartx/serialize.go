package neartx

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/mr-tron/base58"
	"github.com/near/borsh-go"
	"strings"
)

const (
	Binary = iota
	Base64
	Base58
	Hex
	XHex
	ZxHex
	Borsh
)

type Format uint8
type StrFormat string

func (format StrFormat) Serialize(d []byte) ([]byte, error) {
	f := strings.ToLower(string(format))
	switch f {
	case "binary":
		return Format(Binary).Serialize(d)
	case "base64":
		return Format(Base64).Serialize(d)
	case "base58":
		return Format(Base58).Serialize(d)
	case "hex":
		return Format(Hex).Serialize(d)
	case "xhex":
		return Format(XHex).Serialize(d)
	case "zxhex", "0xhex":
		return Format(ZxHex).Serialize(d)
	case "borsh":
		return Format(Borsh).Serialize(d)
	default:
		return nil, fmt.Errorf("neartx: unknown encoding '%s'", format)
	}
}

func (format Format) Serialize(d []byte) ([]byte, error) {
	switch format {
	case Binary:
		return d, nil
	case Base64:
		out := make([]byte, base64.StdEncoding.EncodedLen(len(d)))
		base64.StdEncoding.Encode(out, d)
		return out, nil
	case Base58:
		return []byte(base58.Encode(d)), nil
	case Hex:
		out := make([]byte, hex.EncodedLen(len(d)))
		hex.Encode(out, d)
		return out, nil
	case XHex:
		out := make([]byte, hex.EncodedLen(len(d))+1)
		hex.Encode(out[1:], d)
		out[0] = byte('x')
		return out, nil
	case ZxHex:
		out := make([]byte, hex.EncodedLen(len(d))+2)
		hex.Encode(out[1:], d)
		out[0] = byte('0')
		out[1] = byte('x')
		return out, nil
	case Borsh:
		return borsh.Serialize(d)
	default:
		return nil, fmt.Errorf("neartx: unknown encoding %d", format)
	}
}
