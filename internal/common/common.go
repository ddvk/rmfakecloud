package common

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"hash"
	"hash/crc32"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var signingMethod = jwt.SigningMethodHS256

// GetToken gets the token from the headers
func GetToken(c *gin.Context) (string, error) {
	auth := c.Request.Header["Authorization"]

	if len(auth) < 1 {
		return "", errors.New("missing auth header")
	}
	token := strings.Split(auth[0], " ")
	if len(token) < 2 {
		return "", errors.New("wrong token format")
	}
	strToken := token[1]
	return strToken, nil

}

// ClaimsFromToken parses the claims from the token
func ClaimsFromToken(claim jwt.Claims, token string, key []byte) error {
	_, err := jwt.ParseWithClaims(token, claim,
		func(token *jwt.Token) (interface{}, error) {
			return key, nil
		}, jwt.WithValidMethods([]string{signingMethod.Name}))

	return err
}

// SignClaims signs the claims i.e. creates a token
func SignClaims(claims jwt.Claims, key []byte) (string, error) {
	jwtToken := jwt.NewWithClaims(signingMethod, claims)
	jwtToken.Header["kid"] = "1"
	return jwtToken.SignedString(key)
}

var nameSeparators = regexp.MustCompile(`[./\\]`)

// Sanitize removes all path separators
func Sanitize(param string) string {
	return nameSeparators.ReplaceAllString(param, "")
}

// SanitizeUid
func SanitizeUid(uid string) string {
	return filepath.Clean(filepath.Base(uid))
}

// QueryS sanitize the param
func QueryS(param string, c *gin.Context) string {
	p := c.Query(param)
	return Sanitize(p)
}

// ParamS sanitize the param
func ParamS(param string, c *gin.Context) string {
	p := c.Param(param)
	return Sanitize(p)
}

var table = crc32.MakeTable(crc32.Castagnoli)

func CRC32CWriter() hash.Hash32 {
	// Create a table for CRC32C (Castagnoli polynomial)
	// Create a CRC32C hasher
	return crc32.New(table)
}

func CRC32CSum(crc32c hash.Hash32) string {
	// Compute the CRC32C checksum
	checksum := crc32c.Sum32()

	// Convert the checksum to a byte array
	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, checksum)
	encodedChecksum := base64.StdEncoding.EncodeToString(crcBytes)

	return encodedChecksum
}

func CRC32CFromReader(reader io.Reader) (string, error) {
	crc32c := CRC32CWriter()

	// Copy the reader data into the hasher
	if _, err := io.Copy(crc32c, reader); err != nil {
		return "", err
	}

	return CRC32CSum(crc32c), nil
}

const GCPHashHeader = "x-goog-hash"

func AddHashHeader(c *gin.Context, hash string) {
	c.Header(GCPHashHeader, hash)
}
