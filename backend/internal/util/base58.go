package util

import (
	"fmt"
	"math/big"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var base58Indexes = func() map[rune]int {
	idx := make(map[rune]int, len(base58Alphabet))
	for i, c := range base58Alphabet {
		idx[c] = i
	}
	return idx
}()

// DecodeBase58 decodes a Base58 string to bytes.
func DecodeBase58(s string) ([]byte, error) {
	if s == "" {
		return []byte{}, nil
	}

	result := big.NewInt(0)
	base := big.NewInt(58)

	for _, c := range s {
		v, ok := base58Indexes[c]
		if !ok {
			return nil, fmt.Errorf("invalid base58 character: %q", c)
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(int64(v)))
	}

	decoded := result.Bytes()

	leadingOnes := 0
	for _, c := range s {
		if c == '1' {
			leadingOnes++
		} else {
			break
		}
	}

	if leadingOnes > 0 {
		prefixed := make([]byte, leadingOnes+len(decoded))
		copy(prefixed[leadingOnes:], decoded)
		return prefixed, nil
	}

	return decoded, nil
}
