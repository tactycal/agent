package main

import (
	"encoding/json"
	"os"

	"github.com/tactycal/agent/stubutils"
)

const defaultStatePath = "/var/opt/tactycal/state"

type (
	state struct {
		statePath string
		data      *stateData
	}
	stateData struct {
		Token string
	}
)

func newState(path string) *state {
	return &state{
		statePath: path,
		data:      &stateData{},
	}
}

func (s *state) read() error {
	// try to read existing state
	b, err := stubutils.ReadFile(s.statePath)
	if err != nil {
		return err
	}

	// decode contents
	if err := json.Unmarshal(b, s.data); err != nil {
		return err
	}

	return nil
}

func (s *state) save() error {
	// encode data
	b, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	// write data to file
	if err := stubutils.WriteFile(s.statePath, b, 0600); err != nil {
		return err
	}

	return nil
}

func (s *state) Reset() error {
	return os.Remove(s.statePath)
}

func (s *state) GetToken() (string, error) {
	if err := s.read(); err != nil {
		return "", err
	}

	return s.data.Token, nil
}

func (s *state) SetToken(token string) error {
	s.data.Token = token
	return s.save()
}
