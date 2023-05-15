package serialize

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/mr-tron/base58"
	"github.com/near/borsh-go"
	"strings"
	"bytes"
)

func (format StrFormat) Deserialize(d []byte) ([]byte, error) {
	f := strings.ToLower(string(format))
	switch f {
	case "binary":
		return Format(Binary).Deserialize(d)
	case "base64":
		return Format(Base64).Deserialize(d)
	case "base58":
		return Format(Base58).Deserialize(d)
	case "hex":
		return Format(Hex).Deserialize(d)
	case "xhex":
		return Format(XHex).Deserialize(d)
	case "zxhex", "0xhex":
		return Format(ZxHex).Deserialize(d)
	case "borsh":
		return Format(Borsh).Deserialize(d)
	default:
		return nil, fmt.Errorf("neartx: unknown encoding '%s'", format)
	}
}

func trim(d []byte)[]byte{
	return bytes.Trim(d,"\n ")
}

func (format Format) Deserialize(d []byte) ([]byte, error) {
	switch format {
	case Binary:
		return d, nil
	case Base64:
		d=trim(d)
		out := make([]byte, base64.StdEncoding.DecodedLen(len(d)))
		if _, err := base64.StdEncoding.Decode(out, d); err != nil {
			return nil, err
		}
		return out, nil
	case Base58:
		d=trim(d)
		return base58.Decode(string(d))
	case Hex, XHex, ZxHex:
		d=trim(d)
		if d[0] == 'x' {
			d = d[1:]
		} else if d[0] == '0' && d[1] == 'x' {
			d = d[2:]
		}
		out := make([]byte, hex.DecodedLen(len(d)))
		if _, err := hex.Decode(out, d); err != nil {
			return nil, err
		}
		return out, nil
	case Borsh:
		s := make([]byte, 0, len(d))
		err := borsh.Deserialize(s, d)
		return s, err
	default:
		return nil, fmt.Errorf("neartx: unknown encoding %d", format)
	}
}
