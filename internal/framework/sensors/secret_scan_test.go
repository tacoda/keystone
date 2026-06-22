package sensors

import (
	"bytes"
	"strings"
	"testing"
)

func TestSecretScan_BlocksSensitivePath(t *testing.T) {
	cases := []string{
		".env",
		"foo/.env",
		"foo/.env.production",
		"deploy/credentials.json",
		"secrets/key.pem",
		"vault/whatever.txt",
		"path/to/id_rsa",
		"keys/server.pem",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			var buf bytes.Buffer
			res, err := Run("secret-scan", Context{FilePath: p}, &buf)
			if err != nil {
				t.Fatal(err)
			}
			if !res.Block {
				t.Errorf("expected block for %q, got pass", p)
			}
		})
	}
}

func TestSecretScan_AllowsEnvExample(t *testing.T) {
	var buf bytes.Buffer
	res, err := Run("secret-scan", Context{FilePath: ".env.example"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if res.Block {
		t.Errorf("expected pass for .env.example, got block: %s", res.Message)
	}
}

func TestSecretScan_BlocksCredentialPattern(t *testing.T) {
	body := []byte(`const config = {
  awsAccessKey: "AKIAIOSFODNN7EXAMPLE",
  region: "us-east-1",
};`)
	var buf bytes.Buffer
	res, err := Run("secret-scan", Context{FilePath: "config.js", FileContent: body}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Block {
		t.Errorf("expected block on AWS access key, got pass; out=%q", buf.String())
	}
	if strings.Contains(res.Message, "AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("expected redacted preview, got full match in message: %s", res.Message)
	}
}

func TestSecretScan_PassesCleanContent(t *testing.T) {
	body := []byte(`func main() { fmt.Println("hello") }`)
	var buf bytes.Buffer
	res, err := Run("secret-scan", Context{FilePath: "main.go", FileContent: body}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if res.Block {
		t.Errorf("expected pass on clean content, got block: %s", res.Message)
	}
}

func TestSecretScan_NoContent_AdvisoryPass(t *testing.T) {
	var buf bytes.Buffer
	res, err := Run("secret-scan", Context{}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if res.Block {
		t.Errorf("expected advisory pass on empty context, got block: %s", res.Message)
	}
}

func TestRun_UnknownSensor(t *testing.T) {
	var buf bytes.Buffer
	_, err := Run("not-a-sensor", Context{}, &buf)
	if err == nil {
		t.Fatal("expected ErrUnknownSensor, got nil")
	}
	if _, ok := err.(ErrUnknownSensor); !ok {
		t.Errorf("expected ErrUnknownSensor, got %T", err)
	}
}
