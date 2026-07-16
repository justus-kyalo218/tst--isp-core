package handlers

import (
	"testing"
)

func TestNormalizeSubIspPackagePayload(t *testing.T) {
	payload, err := normalizeSubIspPackagePayload(map[string]interface{}{
		"name":      "Starter",
		"duration":  "7 days",
		"price":     "Ksh. 250",
		"amount":    250,
		"tag":       "5 Mbps",
		"dataCapMB": 15360,
	})
	if err != nil {
		t.Fatalf("normalizeSubIspPackagePayload returned error: %v", err)
	}
	if payload.Name != "Starter" {
		t.Fatalf("expected name Starter, got %q", payload.Name)
	}
	if payload.Amount != 250 {
		t.Fatalf("expected amount 250, got %d", payload.Amount)
	}
	if payload.DataCapMB != 15360 {
		t.Fatalf("expected dataCapMB 15360, got %d", payload.DataCapMB)
	}
}

func TestNormalizeSubIspPackagePayloadRejectsInvalidAmount(t *testing.T) {
	_, err := normalizeSubIspPackagePayload(map[string]interface{}{
		"name":      "Starter",
		"duration":  "7 days",
		"price":     "Ksh. 250",
		"amount":    0,
		"tag":       "5 Mbps",
		"dataCapMB": 15360,
	})
	if err == nil {
		t.Fatal("expected invalid amount error")
	}
}
