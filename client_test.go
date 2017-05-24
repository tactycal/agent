package main

import (
	"agent/packageLookup"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
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

	client := &Client{
		uri:   ts.URL,
		token: "token",
	}

	result, err := client.Authenticate()
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

	client := &Client{uri: ts.URL}

	_, err := client.Authenticate()

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

	client := &Client{uri: ts.URL}

	_, err := client.Authenticate()
	if err == nil {
		t.Errorf("An error was expected")
	}
}

func TestSendPackageList_Ok(t *testing.T) {
	expectedReqBody := &SendPackagesRequestBody{
		Host: &Host{Fqdn: "Fqdn"},
		Package: []*packageLookup.Package{
			&packageLookup.Package{
				Name: "Package",
			},
		},
	}

	// setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "JWT token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&ResponseErrorCode{ErrorCodeInvalidToken})
			return
		}

		reqBody := &SendPackagesRequestBody{}
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
	state := NewState(tmpFile.Name())

	// setup a client
	client := &Client{
		state: state,
		uri:   ts.URL,
		host:  expectedReqBody.Host,
	}

	err := client.SendPackageList(expectedReqBody.Package)
	if err != nil {
		t.Errorf("An error was not expected; got \"%s\"", err.Error())
	}
}

func TestSendPackageList_InvalidToken(t *testing.T) {
	expectedErr := "Token was reported as invalid. Perhaps the host was deleted. New host ID will be assigned to this machine."

	// setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "JWT token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&ResponseErrorCode{ErrorCodeInvalidToken})
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
	state := NewState(tmpFile.Name())

	// setup a client
	client := &Client{
		state: state,
		uri:   ts.URL,
	}

	err := client.SendPackageList([]*packageLookup.Package{})

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
		reqUrl := r.URL.Path

		// 0.1 packages will be submitted with expired token
		if strings.HasSuffix(reqUrl, "/submit") && r.Header.Get("Authorization") == "JWT expired_token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&ResponseErrorCode{ErrorCodeExpiredToken})
			return
		}

		// 0.2 a renew process is expected
		if strings.HasSuffix(reqUrl, "/renew") && r.Header.Get("Authorization") == "Token token" {
			var token Token
			json.NewDecoder(r.Body).Decode(&token)
			if token.Token != "expired_token" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			// return a new token
			token.Token = expectedNewToken
			json.NewEncoder(w).Encode(token)

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
	state := NewState(tmpFile.Name())

	// setup a client
	client := &Client{
		state: state,
		uri:   ts.URL,
		token: "token",
	}

	err := client.SendPackageList([]*packageLookup.Package{})

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
