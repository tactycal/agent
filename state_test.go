package main

import (
	"io/ioutil"
	"os"
	"testing"
)

var testStatePath = "testdata/state"

func TestStateReset_Ok(t *testing.T) {
	tmpFile, _ := ioutil.TempFile("", "tactycal-agent-ut")
	defer os.Remove(tmpFile.Name()) // try to delete just in case

	state := NewState(tmpFile.Name())

	// call reset
	err := state.Reset()

	// check error
	if err != nil {
		t.Errorf("Expected error to be nil; got %v", err)
	}

	// tmp file should be gone
	if _, err := os.Stat(tmpFile.Name()); !os.IsNotExist(err) {
		t.Errorf("Expected file to be removed")
	}
}

func TestStateReset_Err(t *testing.T) {
	// create state with unexisting file
	state := NewState("/should/not/exists")

	err := state.Reset()

	// check error
	if err == nil {
		t.Error("Excected error; to nil")
	}
}

func TestStateGetToken_Ok(t *testing.T) {
	stub := newStubs(t,
		&readFileStub{path: testStatePath, output: []byte(`{"token": "TOKEN"}`)})
	defer stub.Close()

	s := NewState(testStatePath)

	token, err := s.GetToken()

	// check error
	if err != nil {
		t.Errorf("Expected error to be nil; got %v", err)
	}

	// check token
	if token != "TOKEN" {
		t.Errorf("Expected \"TOKEN\"; got \"%s\"", token)
	}
}

func TestStateGetToken_InvalidJson(t *testing.T) {
	stub := newStubs(t,
		&readFileStub{path: testStatePath, output: []byte("How to break JSON?")})
	defer stub.Close()

	s := NewState(testStatePath)

	token, err := s.GetToken()

	// check error
	if err == nil {
		t.Errorf("Expected error not to be nil")
	}

	// check token
	if token != "" {
		t.Errorf("Expected token to be empty; got %v", token)
	}
}

func TestStateGetToken_SomeError(t *testing.T) {
	stub := newStubs(t,
		&readFileStub{path: testStatePath, err: ohNoErr})
	defer stub.Close()

	s := NewState(testStatePath)

	token, err := s.GetToken()

	// check error
	if err != ohNoErr {
		t.Errorf("Expected error %v; got %v", ohNoErr, err)
	}

	// check token
	if token != "" {
		t.Errorf("Expected token to be empty; got %v", token)
	}
}

func TestStateSetToken_Ok(t *testing.T) {
	stub := newStubs(t,
		&writeFileStub{path: testStatePath, data: []byte(`{"Token":"TOKEN"}`), mode: 0600})
	defer stub.Close()

	state := NewState(testStatePath)

	err := state.SetToken("TOKEN")

	// check error
	if err != nil {
		t.Errorf("Expected error to be nil; got %v", err)
	}
}

func TestStateSetToken_Error(t *testing.T) {
	stub := newStubs(t,
		&writeFileStub{err: ohNoErr})
	defer stub.Close()

	state := NewState(testStatePath)
	err := state.SetToken("TOKEN")

	if err == nil {
		t.Errorf("Expected error to be set")
	}
}
