package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/tactycal/agent/packagelookup"
)

func TestAuthenticate_Ok(t *testing.T) {
	expectedResult := "Token"

	testCase := `{"token":"Token"}`

	// setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Header.Get("Authorization") != "Token token" {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		fmt.Fprintln(w, string(testCase))
	}))
	defer ts.Close()

	initLogging(false)

	c := &client{
		uri:   ts.URL,
		token: "token",
	}

	result, err := c.Authenticate()
	if err != nil {
		t.Errorf("An error was not expected; got \"%s\"", err.Error())
	}
	if result != expectedResult {
		t.Errorf("Result \"%s\" doesn't match expected \"%s\"", result, expectedResult)
	}
}

func TestAuthenticate_StatusUnauthorized(t *testing.T) {
	expectedErr := fmt.Sprintf("API returned status code %d, expected 200", http.StatusUnauthorized)

	// setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token token" {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
	}))
	defer ts.Close()

	initLogging(false)

	c := &client{uri: ts.URL}

	_, err := c.Authenticate()

	if err == nil {
		t.Error("Expected err to not be nil")
	}

	if err != nil && !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain \"%s\"; got \"%s\"", expectedErr, err.Error())
	}
}

func TestAuthenticate_CloseConn(t *testing.T) {
	// setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()

	// close client connections
	ts.CloseClientConnections()

	initLogging(false)

	c := &client{uri: ts.URL}

	_, err := c.Authenticate()
	if err == nil {
		t.Errorf("An error was expected")
	}
}

func TestSendPackageList_Ok(t *testing.T) {
	expectedReqBody := &sendPackagesRequestBody{
		Host: &Host{Fqdn: "Fqdn"},
		Package: []*packagelookup.Package{
			&packagelookup.Package{
				Name: "Package",
			},
		},
	}

	// setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "JWT token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&responseErrorCode{errorCodeInvalidToken})
			return
		}

		reqBody := &sendPackagesRequestBody{}
		err := json.NewDecoder(r.Body).Decode(reqBody)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if !reflect.DeepEqual(reqBody.Host, expectedReqBody.Host) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	initLogging(false)

	// setup a state temp file
	tmpFile, _ := ioutil.TempFile("", "client-state")
	tmpFile.Write([]byte(`{"Token":"token"}`))
	defer os.Remove(tmpFile.Name())

	// get a state
	state := newState(tmpFile.Name())

	// setup a client
	c := &client{
		state: state,
		uri:   ts.URL,
		host:  expectedReqBody.Host,
	}

	err := c.SendPackageList(expectedReqBody.Package)
	if err != nil {
		t.Errorf("An error was not expected; got \"%s\"", err.Error())
	}
}

func TestSendPackageList_InvalidToken(t *testing.T) {
	expectedErr := ErrInvalidToken.Error()

	// setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "JWT token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&responseErrorCode{errorCodeInvalidToken})
			return
		}
	}))
	defer ts.Close()

	initLogging(false)

	// setup a state temp file
	tmpFile, _ := ioutil.TempFile("", "client-state")
	tmpFile.Write([]byte(`{"Token":"no token"}`))
	defer os.Remove(tmpFile.Name())

	// get a state
	state := newState(tmpFile.Name())

	// setup a client
	c := &client{
		state: state,
		uri:   ts.URL,
	}

	err := c.SendPackageList([]*packagelookup.Package{})

	if err == nil {
		t.Error("Expected err to not be nil")
	}

	if err != nil && !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain \"%s\"; got \"%s\"", expectedErr, err.Error())
	}
}

func TestSendPackageList_ExpiredToken(t *testing.T) {
	expectedErr := "Token was reported as expired. It has been renewed"
	expectedNewToken := "new_token"

	// setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqURL := r.URL.Path

		// 0.1 packages will be submitted with expired token
		if strings.HasSuffix(reqURL, "/submit") && r.Header.Get("Authorization") == "JWT expired_token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&responseErrorCode{errorCodeExpiredToken})
			return
		}

		// 0.2 a renew process is expected
		if strings.HasSuffix(reqURL, "/renew") && r.Header.Get("Authorization") == "Token token" {
			var tkn token
			json.NewDecoder(r.Body).Decode(&tkn)
			if tkn.Token != "expired_token" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			// return a new token
			tkn.Token = expectedNewToken
			json.NewEncoder(w).Encode(tkn)

			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	initLogging(false)

	// setup a state temp file
	tmpFile, _ := ioutil.TempFile("", "client-state")
	tmpFile.Write([]byte(`{"Token":"expired_token"}`))
	defer os.Remove(tmpFile.Name())

	// get a state
	state := newState(tmpFile.Name())

	// setup a client
	c := &client{
		state: state,
		uri:   ts.URL,
		token: "token",
	}

	err := c.SendPackageList([]*packagelookup.Package{})

	if err == nil {
		t.Error("Expected err to not be nil")
	}

	if err != nil && !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain \"%s\"; got \"%s\"", expectedErr, err.Error())
	}

	// new token is expected
	newToken, _ := state.GetToken()
	if newToken != expectedNewToken {
		t.Errorf("Expected token \"%s\"; got \"%s\"", expectedNewToken, newToken)
	}
}
