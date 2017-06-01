package main

import (
	"encoding/json"
	"os"

	"github.com/tactycal/agent/stubUtils"
)

const DefaultStatePath = "/var/opt/tactycal/state"

type (
	State struct {
		statePath string
		data      *StateData
	}
	StateData struct {
		Token string
	}
)

func NewState(path string) *State {
	return &State{
		statePath: path,
		data:      &StateData{},
	}
}

func (s *State) read() error {
	// try to read existing state
	b, err := stubUtils.ReadFile(s.statePath)
	if err != nil {
		return err
	}

	// decode contents
	if err := json.Unmarshal(b, s.data); err != nil {
		return err
	}

	return nil
}

func (s *State) save() error {
	// encode data
	b, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	// write data to file
	if err := stubUtils.WriteFile(s.statePath, b, 0600); err != nil {
		return err
	}

	return nil
}

func (s *State) Reset() error {
	return os.Remove(s.statePath)
}

func (s *State) GetToken() (string, error) {
	if err := s.read(); err != nil {
		return "", err
	}

	return s.data.Token, nil
}

func (s *State) SetToken(token string) error {
	s.data.Token = token
	return s.save()
}
