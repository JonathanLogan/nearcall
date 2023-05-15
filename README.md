# nearcall
Make calls to NEAR contracts with a single argument sourced from file.

## Installation

Requires go >= 1.20.

Install: `go install github.com/JonathanLogan/nearcall/cmd/nearcall@0.0.4`

## Usage

Run `nearcall --help` for information.

Returns execution errors in original format (exit code 3) on stderr. On success no data is printed (exit code 0), unless "-v" flag is given.

## Credentials/Keys

nearcall reads account keys from $HOME/.near-credentials/$NETWORK/$ACCOUNT.json, like near-cli.
Hardware wallets are NOT supported.


## Supported encodings:

  - Binary (as is)
  - Borsh 
  - Base64 (with padding)
  - Base58 (near compatible)
  - Hex (lowercase hexadecimal)
  - xHex (lowercase hexadecimal with leading "x")
  - ZxHex (lowercase hexadecimal with leading "0x")
