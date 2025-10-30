package consumer1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"model"
	"net/http"
	"net/url"
)

var (
	// ErrNotFound represents a resource not found (404)
	ErrNotFound = errors.New("not found")
)

type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

// GetUsers gets all users from the API
func (c *Client) GetProducts() ([]model.Product, error) {
	req, err := c.newRequest("GET", "/api/v1/products", nil)
	if err != nil {
		return nil, err
	}
	var products []model.Product
	_, err = c.do(req, &products)

	return products, err
}

// GetUser gets a user by ID from the API
func (c *Client) GetProduct(id int) (*model.Product, error) {
	req, err := c.newRequest("GET", fmt.Sprintf("/api/v1/products/%d", id), nil)
	if err != nil {
		return nil, err
	}
	var product model.Product
	res, err := c.do(req, &product)
	if err != nil {
		switch res.StatusCode {
		case http.StatusNotFound:
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

// NewClient creates a new API client with the given base URL and HTTP client.
func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
