// simhash.go: Computes SimHash for HTTP response bodies, used for clustering similar responses.
package islazy

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/fnv"
	"strings"
	"unicode"
)

const responseSimHashPrefix = "s:"

// NormalizeResponseBody normalizes the response body for easier similarity comparison
func NormalizeResponseBody(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	var v interface{}
	if json.Unmarshal([]byte(body), &v) == nil {
		compact, err := json.Marshal(v)
		if err == nil {
			return string(compact)
		}
	}

	return strings.Join(strings.Fields(body), " ")
}

// ResponseSimHash calculates the simhash (8 bytes) of the response body
func ResponseSimHash(body string) []byte {
	normalized := NormalizeResponseBody(body)
	if normalized == "" {
		return nil
	}

	hash := simHash64(tokenize(normalized))
	out := make([]byte, 8)
	binary.BigEndian.PutUint64(out, hash)
	return out
}

// FormatResponseSimHash formats the simhash into a database storage string s:<hex>
func FormatResponseSimHash(hash []byte) string {
	if len(hash) == 0 {
		return ""
	}
	return responseSimHashPrefix + hex.EncodeToString(hash)
}

// ParseResponseSimHash parses a string in the format s:<hex>
func ParseResponseSimHash(hashStr string) ([]byte, error) {
	if hashStr == "" {
		return nil, nil
	}
	if !strings.HasPrefix(hashStr, responseSimHashPrefix) {
		return nil, errors.New("invalid response simhash format: missing 's:' prefix")
	}
	return hex.DecodeString(strings.TrimPrefix(hashStr, responseSimHashPrefix))
}

func tokenize(text string) []string {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.ToLower(strings.TrimSpace(f))
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func simHash64(tokens []string) uint64 {
	if len(tokens) == 0 {
		return 0
	}

	var weights [64]int
	for _, token := range tokens {
		h := fnv.New64a()
		_, _ = h.Write([]byte(token))
		value := h.Sum64()
		for i := 0; i < 64; i++ {
			if (value>>i)&1 == 1 {
				weights[i]++
			} else {
				weights[i]--
			}
		}
	}

	var hash uint64
	for i := 0; i < 64; i++ {
		if weights[i] > 0 {
			hash |= 1 << i
		}
	}
	return hash
}
