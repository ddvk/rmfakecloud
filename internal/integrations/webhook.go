package integrations

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
)

type Webhook struct {
	Endpoint string
}

func newWebhook(i model.IntegrationConfig) *Webhook {
	return &Webhook{
		Endpoint: i.Endpoint,
	}
}

func (i *Webhook) SendMessage(data messages.IntegrationMessageData, img image.Image) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add data field
	if mdata, err := json.Marshal(data); err != nil {
		return "", err
	} else if err := writer.WriteField("data", string(mdata)); err != nil {
		return "", err
	}

	// Add attachment
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	part, err := writer.CreateFormFile("attachment", "reMarkable.png")
	if err != nil {
		return "", err
	}
	if _, err := part.Write(buf.Bytes()); err != nil {
		return "", err
	}

	// Close
	if err := writer.Close(); err != nil {
		return "", err
	}

	// Do the request
	resp, err := http.Post(i.Endpoint, writer.FormDataContentType(), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(responseData), nil
}
