package config

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func TestTLSPathsRequireCompletePair(t *testing.T) {
	for _, test := range []struct {
		name, cert, key string
		valid           bool
	}{
		{"neither", "", "", true},
		{"both", "/path/to/cert", "/path/to/key", true},
		{"missing-key", "/path/to/cert", "", false},
		{"missing-cert", "", "/path/to/key", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validateTLSPaths(test.cert, test.key)
			if (err == nil) != test.valid {
				t.Fatalf("unexpected result for cert=%q key=%q: %v", test.cert, test.key, err)
			}
		})
	}
}

// FromEnv stops the process with log.Fatal, so the startup check runs in a
// child process: the test binary re-runs only this test with a marker set.
func TestFromEnvRefusesCertWithoutKey(t *testing.T) {
	if os.Getenv("RMFAKECLOUD_TEST_PARTIAL_TLS") == "1" {
		os.Setenv(envTLSCert, "/path/to/cert")
		os.Unsetenv(envTLSKey)
		os.Setenv(envDataDir, t.TempDir())
		FromEnv()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestFromEnvRefusesCertWithoutKey$")
	cmd.Env = append(os.Environ(), "RMFAKECLOUD_TEST_PARTIAL_TLS=1")
	out, err := cmd.CombinedOutput()
	if _, exited := err.(*exec.ExitError); !exited {
		t.Fatalf("FromEnv started with only %s set (err=%v); output:\n%s", envTLSCert, err, out)
	}
	if !bytes.Contains(out, []byte("must be set together")) {
		t.Fatalf("FromEnv stopped without the expected message; output:\n%s", out)
	}
}
