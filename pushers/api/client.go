package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	Url   string
	Token string
}

func New(config *Config) *Client {
	baseURL, _ := url.Parse(config.Url)

	hc := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 5,
		},
		Timeout: time.Duration(20) * time.Second,
	}

	c := &Client{client: hc, BaseURL: baseURL, Token: config.Token}
	return c
}

type Client struct {
	client  *http.Client
	BaseURL *url.URL
	Token   string
}

type Response struct {
	*http.Response
}

func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	var err error
	var req *http.Request

	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	if req, err = http.NewRequest(method, u.String(), buf); err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", c.Token))
	}

	req.Header.Set("Content-Type", "application/json")
	return req, err
}

// Do sends an API request and returns the API response.  The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred.  If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	var err error
	var resp *http.Response

	if resp, err = c.client.Do(req); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if err = CheckResponse(resp); err != nil {
		return resp, err
	}

	if v == nil {
		return resp, nil
	}

	if w, ok := v.(io.Writer); ok {
		io.Copy(w, resp.Body)
	} else {
		err = json.NewDecoder(resp.Body).Decode(v)
	}

	return resp, err
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}

type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Code     string         `json:"code"`    // error message
	Message  string         `json:"message"` // error message
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%s (%d)",
		r.Message, r.Response.StatusCode)
}
