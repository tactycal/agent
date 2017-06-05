package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tactycal/agent/packageLookup"
)

const (
	DefaultClientTimeout  = time.Second * 3
	ErrorCodeExpiredToken = "TOKEN_EXPIRED"
	ErrorCodeInvalidToken = "TOKEN_INVALID"
	apiVersionPrefix      = "v2"
)

type Client struct {
	token    string
	host     *Host
	uri      string
	proxyUrl *url.URL
	state    *State
	timeout  time.Duration
}

type SendPackagesRequestBody struct {
	*Host
	Package []*packageLookup.Package `json:"packages"`
}

type ResponseErrorCode struct {
	Error string `json:"error"`
}

type Token struct {
	Token string `json:"token"`
}

func NewClient(cfg *Config, host *Host, state *State, timeout time.Duration) *Client {
	// copy labels from config to host
	host.Labels = cfg.Labels

	// compose the client
	return &Client{
		token:    cfg.Token,
		host:     host,
		uri:      cfg.Uri,
		proxyUrl: cfg.Proxy,
		state:    state,
		timeout:  timeout,
	}
}

func (c *Client) Authenticate() (string, error) {
	// create a request
	rsp, err := c.apiRequest("POST", "/agent/auth", fmt.Sprintf("Token %s", c.token), &c.host)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	log.Debugf("Got a %d response", rsp.StatusCode)

	// validate the response code
	if rsp.StatusCode != 200 {
		return "", fmt.Errorf("API returned status code %d, expected 200", rsp.StatusCode)
	}

	// decode the response
	var rspData Token
	decoder := json.NewDecoder(rsp.Body)
	err = decoder.Decode(&rspData)
	if err != nil {
		return "", err
	}

	return rspData.Token, nil
}

func (c *Client) SendPackageList(packages []*packageLookup.Package) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	body := &SendPackagesRequestBody{
		Host:    c.host,
		Package: packages,
	}

	// create a request
	rsp, err := c.apiRequest("POST", "/agent/submit", fmt.Sprintf("JWT %s", token), body)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	log.Debugf("Got a %d response", rsp.StatusCode)

	if rsp.StatusCode == http.StatusNoContent {

		// check if a new token (header X-Token) was returned by the API; if so,
		// update the state
		if h := rsp.Header.Get("X-Token"); h != "" {
			log.Debug("Received a new token from the API; updating state")
			c.state.SetToken(h)
		}

		return nil
	}

	// handle invalid or expired token response
	if rsp.StatusCode == http.StatusUnauthorized {
		// check error code
		errCode := &ResponseErrorCode{}
		if err := json.NewDecoder(rsp.Body).Decode(errCode); err != nil {
			return err
		}

		if errCode.Error == ErrorCodeInvalidToken {
			if err := c.state.Reset(); err != nil {
				return err
			}
			return fmt.Errorf("Token was reported as invalid. Perhaps the host was deleted. New host ID will be assigned to this machine.")
		}

		// renew a token
		if errCode.Error == ErrorCodeExpiredToken {
			err := c.renewToken(token)
			if err == nil {
				err = fmt.Errorf("Token was reported as expired. It has been renewed")
			}
			return err
		}
	}

	return fmt.Errorf("API returned status code %d, expected 204", rsp.StatusCode)
}

func (c *Client) getToken() (string, error) {
	// try to read token from state
	token, err := c.state.GetToken()
	if err == nil && token != "" {
		return token, nil
	}

	// get a new token from API
	log.Debug("Agent not authenticated yet.")
	token, err = c.Authenticate()
	if err != nil {
		return "", err
	}

	// write the token new token to state
	if err := c.state.SetToken(token); err != nil {
		return "", err
	}

	return token, nil
}

func (c *Client) renewToken(token string) error {
	// create a request
	rsp, err := c.apiRequest("POST", "/agent/renew", fmt.Sprintf("Token %s", c.token), &Token{token})
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	log.Debugf("Got a %d response", rsp.StatusCode)

	if rsp.StatusCode == 401 {
		c.state.Reset()
		return fmt.Errorf("Token was reported as invalid. Perhaps the host was deleted. New host ID will be assigned to this machine.")
	}

	if rsp.StatusCode != 200 {
		return fmt.Errorf("Token could not be renewed")
	}

	var rspData Token
	decoder := json.NewDecoder(rsp.Body)
	err = decoder.Decode(&rspData)
	if err != nil {
		return err
	}

	// write the new token to state
	err = c.state.SetToken(rspData.Token)

	return err
}

func (c *Client) apiRequest(method, endpoint, authorization string, input interface{}) (*http.Response, error) {
	// encode body
	body := bytes.NewBuffer(nil)
	if input != nil {
		enc := json.NewEncoder(body)
		if err := enc.Encode(input); err != nil {
			return nil, err
		}
	}

	// strip slashes from the beginning of the endpoint
	endpoint = strings.TrimLeft(endpoint, "/")

	// build the request
	uri := fmt.Sprintf("%s/%s/%s", c.uri, apiVersionPrefix, endpoint)
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	// execute the request
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(c.proxyUrl),
		},
		Timeout: c.timeout,
	}

	log.Debugf("Sending a %s request to %s", req.Method, req.URL)

	return client.Do(req)
}
