package ddblocal

import "crypto/rand"

type randStringGenFunc func() (string, error)

func (r randStringGenFunc) Generate() (string, error) {
	return r()
}

func genRandomString() (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, 12)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

func NewStringGenerator() StringGenerator {
	return randStringGenFunc(genRandomString)
}
