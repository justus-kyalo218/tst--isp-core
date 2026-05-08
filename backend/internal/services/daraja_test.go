package services

import (
	"net/http/httptest"
	"testing"
)

func TestResolveCallbackURLAcceptsConfiguredHTTPSURL(t *testing.T) {
	t.Parallel()

	got, err := resolveCallbackURL("https://billing.example.com/api/mpesa/callback", nil)
	if err != nil {
		t.Fatalf("resolveCallbackURL returned error: %v", err)
	}
	if got != "https://billing.example.com/api/mpesa/callback" {
		t.Fatalf("unexpected callback url: %s", got)
	}
}

func TestResolveCallbackURLRejectsLocalhost(t *testing.T) {
	t.Parallel()

	if _, err := resolveCallbackURL("http://localhost:8080/api/mpesa/callback", nil); err == nil {
		t.Fatal("expected localhost callback url to be rejected")
	}
}

func TestResolveCallbackURLInfersForwardedHTTPSHost(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("POST", "http://127.0.0.1/api/mpesa/stkpush", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "isp-demo.ngrok-free.app")

	got, err := resolveCallbackURL("", req)
	if err != nil {
		t.Fatalf("resolveCallbackURL returned error: %v", err)
	}
	if got != "https://isp-demo.ngrok-free.app/api/mpesa/callback" {
		t.Fatalf("unexpected callback url: %s", got)
	}
}
