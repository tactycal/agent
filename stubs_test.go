package main

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var ohNoErr = errors.New("oh no")

type (
	Stubs struct {
		queue           []ioStub
		t               *testing.T
		origExecCommand func(string, ...string) ([]byte, error)
		origReadFile    func(string) ([]byte, error)
		origWriteFile   func(string, []byte, os.FileMode) error
	}

	ioStub interface {
		check(*testing.T, []string) ([]byte, error)
	}

	cmdStub struct {
		cmd      string
		args     []string
		stubFile string
		output   []byte
		err      error
	}

	readFileStub struct {
		path     string
		stubFile string
		output   []byte
		err      error
	}

	writeFileStub struct {
		path string
		data []byte
		mode os.FileMode
		err  error
	}
)

func (c *cmdStub) check(t *testing.T, in []string) ([]byte, error) {
	// check number of parameters
	if len(in) < 1 {
		t.Fatalf("Command not defined")
	}

	// check command
	if c.cmd != "" && c.cmd != in[0] {
		t.Errorf("Expected command \"%s\"; received \"%s\"", c.cmd, in[0])
	}

	// check arguments
	if len(c.args) > 0 {
		if len(in) <= 1 {
			t.Errorf("Expected some arguments %+v; got none", c.args)
		} else if !reflect.DeepEqual(c.args, in[1:]) {
			t.Errorf("Expected arguments \"%+v\"; received \"%+v\"", c.args, in[1:])
		}
	}

	// open the stub file if specified
	if c.stubFile != "" {
		b, err := ioutil.ReadFile(c.stubFile)
		if err != nil {
			t.Fatalf("Failed to read stub file %s; err = %v", c.stubFile, err)
		}
		c.output = b
	}

	// return stubbed data
	return c.output, c.err
}

func (r *readFileStub) check(t *testing.T, in []string) ([]byte, error) {
	if len(in) != 1 {
		t.Fatal("Expected one argument to be passed to readFile; got none")
	}

	path := in[0]

	// check requested path
	if r.path != "" && r.path != path {
		t.Errorf("Expected readFile's requested path to match \"%s\"; got \"%s\"", r.path, path)
	}

	// open the stub file if specified
	if r.stubFile != "" {
		b, err := ioutil.ReadFile(r.stubFile)
		if err != nil {
			t.Fatalf("Failed to read stub file %s; err = %v", r.stubFile, err)
		}
		r.output = b
	}

	// return stubbed response
	return r.output, r.err
}

func (w *writeFileStub) check(t *testing.T, in []string) ([]byte, error) {
	// check path
	if w.path != "" && w.path != in[0] {
		t.Errorf("Expected path to equal \"%s\"; got \"%s\"", w.path, in[0])
	}

	// check contents
	if len(w.data) > 0 && !reflect.DeepEqual(w.data, []byte(in[1])) {
		t.Errorf("Excpected data to equal %v; got %v", w.data, in[1])
	}

	// check mode
	if w.mode != 0 && string(w.mode) != in[2] {
		t.Errorf("Expected mode %v", w.mode)
	}

	return nil, w.err
}

// Creates a new stub collection and stubs the commands.
func newStubs(t *testing.T, stubs ...ioStub) *Stubs {
	s := &Stubs{
		t:               t,
		queue:           stubs,
		origExecCommand: execCommand,
		origReadFile:    readFile,
		origWriteFile:   writeFile,
	}

	readFile = func(path string) ([]byte, error) {
		stub := s.Get()

		// did we get something
		if stub == nil {
			t.Fatalf("Stub queue is already empty when calling readFile(\"%s\")", path)
		}

		// validate the type
		if _, ok := stub.(*readFileStub); !ok {
			t.Fatalf("Expected a readFile stub; got %T", stub)
		}

		// check it
		return stub.check(t, []string{path})
	}

	execCommand = func(cmd string, args ...string) ([]byte, error) {
		stub := s.Get()

		// did we get something
		if stub == nil {
			t.Fatalf("Stub queue is already empty when calling execCommand(\"%s\", %+v)", cmd, args)
		}

		// validate the type
		if _, ok := stub.(*cmdStub); !ok {
			t.Fatalf("Expected a execCommand stub; got %T", stub)
		}

		// check it
		return stub.check(t, append([]string{cmd}, args...))
	}

	writeFile = func(filename string, data []byte, perm os.FileMode) error {
		stub := s.Get()

		// did we get something
		if stub == nil {
			t.Fatalf("Stub queue is already empty when calling writeFile(\"%s\" ...)", filename)
		}

		// validate the type
		if _, ok := stub.(*writeFileStub); !ok {
			t.Fatalf("Expected a writeFile stub; got %T", stub)
		}

		// check it
		_, err := stub.check(t, []string{filename, string(data), string(perm)})
		return err
	}

	return s
}

// Add a single stub to end of the queue
func (s *Stubs) Add(stub ioStub) {
	s.queue = append(s.queue, stub)
}

// Get pops the first item from the queue. Returns nil if the queue is already empty.
func (s *Stubs) Get() ioStub {
	if len(s.queue) == 0 {
		return nil
	}

	first := s.queue[0]
	s.queue = s.queue[1:]
	return first
}

// Resets stubbed commands and and checks if the queue is empty.
func (s *Stubs) Close() {
	// reset functions
	execCommand = s.origExecCommand
	readFile = s.origReadFile
	writeFile = s.origWriteFile

	if len(s.queue) > 0 {
		s.t.Errorf("Excpected the queue to be empty. %d items left in the queue", len(s.queue))
		for _, q := range s.queue {
			s.t.Errorf("Left in queue: %+v", q)
		}
	}
}
