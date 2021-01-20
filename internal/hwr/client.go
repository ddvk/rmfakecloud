package hwr

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

var key, hmackey string

func init() {
	key = os.Getenv("RMAPI_HWR_APPLICATIONKEY")
	if key == "" {
		log.Println("if you want HWR, provide the myScript applicationKey in: RMAPI_HWR_APPLICATIONKEY")
	}
	hmackey = os.Getenv("RMAPI_HWR_HMAC")
	if hmackey == "" {
		log.Println("provide the myScript hmac in: RMAPI_HWR_HMAC")
	}
}

const url = "https://cloud.myscript.com/api/v4.0/iink/batch"
const JIIX = "application/vnd.myscript.jiix"

func SendRequest(data []byte) (body []byte, err error) {
	if key == "" || hmackey == "" {
		return nil, fmt.Errorf("no hwr key set")
	}
	fullkey := key + hmackey
	mac := hmac.New(sha512.New, []byte(fullkey))
	mac.Write(data)
	result := hex.EncodeToString(mac.Sum(nil))

	client := http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Set("Accept", JIIX)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("applicationKey", key)
	req.Header.Set("hmac", result)

	res, err := client.Do(req)

	if err != nil {
		return
	}
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Not ok, Status: %d", res.StatusCode)
		return
	}

	return body, nil
}
