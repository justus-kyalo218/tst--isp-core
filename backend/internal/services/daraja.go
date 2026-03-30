package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type DarajaClient struct {
	AuthURL         string
	STKURL          string
	ConsumerKey     string
	ConsumerSecret  string
	Shortcode       string
	Passkey         string
	CallbackURL     string
	AccountRef      string
	TransactionDesc string
	TransactionType string
	Timeout         time.Duration
}

type STKPushRequest struct {
	Phone       string
	Amount      int
	PackageName string
}

type STKPushResult struct {
	Message           string
	Phone             string
	Amount            int
	Timestamp         string
	CheckoutRequestID string
	MerchantRequestID string
}

func NewDarajaFromEnv() (*DarajaClient, error) {
	client := &DarajaClient{
		AuthURL:         strings.TrimSpace(os.Getenv("DARAJA_AUTH_URL")),
		STKURL:          strings.TrimSpace(os.Getenv("DARAJA_STK_URL")),
		ConsumerKey:     strings.TrimSpace(os.Getenv("DARAJA_CONSUMER_KEY")),
		ConsumerSecret:  strings.TrimSpace(os.Getenv("DARAJA_CONSUMER_SECRET")),
		Shortcode:       strings.TrimSpace(os.Getenv("DARAJA_SHORTCODE")),
		Passkey:         strings.TrimSpace(os.Getenv("DARAJA_PASSKEY")),
		CallbackURL:     strings.TrimSpace(os.Getenv("DARAJA_CALLBACK_URL")),
		AccountRef:      strings.TrimSpace(os.Getenv("DARAJA_ACCOUNT_REF")),
		TransactionDesc: strings.TrimSpace(os.Getenv("DARAJA_TRANSACTION_DESC")),
		TransactionType: strings.TrimSpace(os.Getenv("DARAJA_TRANSACTION_TYPE")),
		Timeout:         12 * time.Second,
	}

	if client.AuthURL == "" || client.STKURL == "" {
		return nil, errors.New("missing Daraja URLs in env (DARAJA_AUTH_URL, DARAJA_STK_URL)")
	}
	if client.ConsumerKey == "" || client.ConsumerSecret == "" || client.Shortcode == "" || client.Passkey == "" {
		return nil, errors.New("missing Daraja credentials in env")
	}
	if client.CallbackURL == "" || client.AccountRef == "" || client.TransactionDesc == "" {
		return nil, errors.New("missing Daraja business fields in env")
	}
	if client.TransactionType == "" {
		client.TransactionType = "CustomerPayBillOnline"
	}

	return client, nil
}

func (c *DarajaClient) STKPush(req STKPushRequest) (*STKPushResult, error) {
	phone, err := normalizePhone(req.Phone)
	if err != nil {
		return nil, err
	}

	accessToken, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102150405")
	password := base64.StdEncoding.EncodeToString([]byte(c.Shortcode + c.Passkey + timestamp))

	payload := map[string]interface{}{
		"BusinessShortCode": c.Shortcode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   c.TransactionType,
		"Amount":            req.Amount,
		"PartyA":            phone,
		"PartyB":            c.Shortcode,
		"PhoneNumber":       phone,
		"CallBackURL":       c.CallbackURL,
		"AccountReference":  c.AccountRef,
		"TransactionDesc":   c.TransactionDesc,
	}

	if req.PackageName != "" {
		payload["AccountReference"] = req.PackageName
	}

	body, _ := json.Marshal(payload)
	httpClient := &http.Client{Timeout: c.Timeout}
	httpReq, err := http.NewRequest(http.MethodPost, c.STKURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stk push failed: %v", respBody)
	}

	checkoutID, _ := respBody["CheckoutRequestID"].(string)
	merchantID, _ := respBody["MerchantRequestID"].(string)

	return &STKPushResult{
		Message:           "STK push initiated",
		Phone:             req.Phone,
		Amount:            req.Amount,
		Timestamp:         time.Now().Format(time.RFC3339),
		CheckoutRequestID: checkoutID,
		MerchantRequestID: merchantID,
	}, nil
}

func (c *DarajaClient) getAccessToken() (string, error) {
	httpClient := &http.Client{Timeout: c.Timeout}
	req, err := http.NewRequest(http.MethodGet, c.AuthURL, nil)
	if err != nil {
		return "", err
	}
	basic := base64.StdEncoding.EncodeToString([]byte(c.ConsumerKey + ":" + c.ConsumerSecret))
	req.Header.Set("Authorization", "Basic "+basic)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("auth failed: %s", resp.Status)
	}

	var data struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.AccessToken == "" {
		return "", errors.New("auth token missing in response")
	}
	return data.AccessToken, nil
}

func normalizePhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	phone = strings.TrimPrefix(phone, "+")

	if strings.HasPrefix(phone, "07") && len(phone) == 10 {
		return "254" + phone[1:], nil
	}
	if strings.HasPrefix(phone, "2547") && len(phone) == 12 {
		return phone, nil
	}
	return "", errors.New("invalid phone number format")
}
