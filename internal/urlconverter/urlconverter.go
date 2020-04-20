package urlconverter

import (
	"fmt"
)

var elems = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// IDToShortURL generates short URL based on integer id.
func IDToURL(id int64) []byte {
	var url []byte

	for id > 0 {
		url = append(url, elems[id%62])
		id /= 62
	}

	return url
}

// URLToID restores id from short URL.
func URLToID(url []byte) (int64, error) {
	var id int64

	for _, b := range url {
		switch {
		case 'a' <= b && b <= 'z':
			id = id*62 + int64(b-'a')

		case 'A' <= b && b <= 'Z':
			id = id*62 + int64(b-'A') + 26

		case '0' <= b && b <= '9':
			id = id*62 + int64(b-'0') + 52

		default:
			return id, fmt.Errorf("URL contains unsupported symbol %q", b)
		}
	}

	return id, nil
}
