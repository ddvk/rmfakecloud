package integrations

import (
	"fmt"
	"os"
	"testing"

	"github.com/zgs225/rmfakecloud/internal/model"
	"github.com/sirupsen/logrus"
)

var skip = false

func init() {
	if os.Getenv("WEBDAV_ADDRESS") == "" {
		fmt.Println("Skipping tests, no webdav")
		skip = true
	}
	logrus.SetLevel(logrus.TraceLevel)

}
func TestListFiles(t *testing.T) {
	if skip {
		t.Skip()
	}

	w := GetWebDav()
	res, err := w.List("root", 2)
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}
func TestUpload(t *testing.T) {
	if skip {
		t.Skip()
	}

	w := GetWebDav()

	file, err := os.Open("test.md")
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()

	id, err := w.Upload("root", "test", "md", file)
	if err != nil {
		t.Error(err)
	}
	t.Log(id)
}

func GetWebDav() IntegrationProvider {
	config := model.IntegrationConfig{
		Address:  os.Getenv("WEBDAV_ADDRESS"),
		Username: os.Getenv("WEBDAV_USERNAME"),
		Password: os.Getenv("WEBDAV_PASSWORD"),
		Insecure: true,
	}
	return newWebDav(config)
}
