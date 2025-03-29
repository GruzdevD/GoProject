package api

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type APIClient struct {
    BaseURL    string
    HTTPClient *http.Client
}

type Message struct {
    Content string `json:"content"`
}

func NewAPIClient(baseURL string) *APIClient {
    return &APIClient{
        BaseURL:    baseURL,
        HTTPClient: &http.Client{},
    }
}

func (c *APIClient) SendMessage(msg Message) error {
    jsonData, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", c.BaseURL+"/messages", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
    }

    return nil
}