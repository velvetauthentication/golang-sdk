package vauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type VAuth struct {
	httpClient *http.Client
	apiBaseURL string
	appID      string
	secret     string
	version    string
	hwid       string
}

func NewVAuth(appID, secret, version string) *VAuth {
	return &VAuth{
		httpClient: &http.Client{},
		apiBaseURL: "https://velvetauth.com/api/",
		appID:      appID,
		secret:     secret,
		version:    version,
		hwid:       "<get_hwid_here>",
	}
}

func (va *VAuth) Post(endpoint string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", va.apiBaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return va.httpClient.Do(req)
}

func (va *VAuth) Init() (bool, error) {
	requestData := map[string]interface{}{
		"type":    "init",
		"app_id":  va.appID,
		"secret":  va.secret,
		"version": va.version,
	}

	response, err := va.Post("index.php", requestData)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Initialization failed: %d", response.StatusCode)
	}

	var jsonResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		return false, err
	}

	if jsonResponse != nil && jsonResponse["error"] != nil && jsonResponse["error"].(string) == "wrong_version" {
		downloadURL := jsonResponse["download_url"].(string)
		fmt.Println("Your are using an outdated version of the program. Redirecting to update URL:", downloadURL)
		return false, nil
	}

	return true, nil
}

func (va *VAuth) RegisterLicense(username, password, licenseKey, email string) (bool, error) {
	requestData := map[string]interface{}{
		"type":        "register",
		"app_id":      va.appID,
		"username":    username,
		"password":    password,
		"hwid":        va.hwid,
		"license_key": licenseKey,
		"email":       email,
	}

	response, err := va.Post("index.php", requestData)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Registration failed: %d", response.StatusCode)
	}

	var jsonResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		return false, err
	}

	if jsonResponse != nil && jsonResponse["message"] != nil && jsonResponse["message"].(string) == "License registered successfully" {
		return true, nil
	}

	errorMessage := "Unknown error"
	if jsonResponse != nil && jsonResponse["error"] != nil {
		errorMessage = jsonResponse["error"].(string)
	}

	return false, fmt.Errorf("Registration failed: %s", errorMessage)
}
