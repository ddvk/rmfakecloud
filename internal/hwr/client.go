package hwr

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ddvk/rmfakecloud/internal/config"
)

const (
	url = "https://cloud.myscript.com/api/v4.0/iink/batch"

	// JIIX jiix type
	JIIX = "application/vnd.myscript.jiix"
)

type HWRClient struct {
	Cfg *config.Config
}

func DoLangOverride(originalData []byte, overrideLang string) ([]byte, error) {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(originalData, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse json", err)
	}

	config, ok := jsonData["configuration"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("configuration schema missing in json")
	}

	config["lang"] = overrideLang

	modifiedData, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate modified json", err)
	}

	return modifiedData, nil
}

// SendRequest sends the request
func (hwr *HWRClient) SendRequest(data []byte) (body []byte, err error) {
	if hwr.Cfg.HWRLangOverride != "" {
		overrideLang := hwr.Cfg.HWRLangOverride
		modifiedData, err := DoLangOverride(data, overrideLang)
		if err != nil {
			return nil, fmt.Errorf("failed to override language", err)
		}
		data = modifiedData		
	}

	if hwr.Cfg == nil || hwr.Cfg.HWRApplicationKey == "" || hwr.Cfg.HWRHmac == "" {
		return nil, fmt.Errorf("no hwr key set")
	}
	appKey := hwr.Cfg.HWRApplicationKey
	fullkey := appKey + hwr.Cfg.HWRHmac
	mac := hmac.New(sha512.New, []byte(fullkey))
	mac.Write(data)
	result := hex.EncodeToString(mac.Sum(nil))

	client := http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", JIIX)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("applicationKey", appKey)
	req.Header.Set("hmac", result)

	res, err := client.Do(req)

	if err != nil {
		return
	}
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("not ok, Status: %d", res.StatusCode)
		return
	}

	return body, nil
}
