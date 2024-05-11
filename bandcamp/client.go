package bandcamp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type userTransport struct {
	cookie string
	base   http.RoundTripper
}

func (ut *userTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Cookie", ut.cookie)
	return ut.base.RoundTrip(req)
}

type Client struct {
	c *http.Client
}

type APIError struct {
	IsError bool   `json:"error"`
	Message string `json:"error_message"`
}

func (e APIError) Error() string {
	return e.Message
}

func New(bandcampCookie string) *Client {
	c := &http.Client{
		Transport: &userTransport{cookie: bandcampCookie, base: http.DefaultTransport},
	}

	return &Client{
		c: c,
	}
}

func (c *Client) Request(method, endpoint string, body, data interface{}) error {
	buf := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, bc+"/api/"+endpoint, buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	e := new(APIError)
	if err := json.Unmarshal(b, &e); err == nil {
		if e.IsError {
			return e
		}
	}

	if data != nil {
		return json.Unmarshal(b, &data)
	}

	return nil
}
