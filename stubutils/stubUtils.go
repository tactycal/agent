// Package stubutils provides wrapper functions for reading from file, writing
// to file, executing a command and stub interface for unit testing.
//
// A stubutils package has predefined structs for mocking the wrapper functions
// which has been mentioned above. All those sturcts implements ioStub interface
// which will inform you through tester interface if any of the passing
// arguments to functions or their outputs will not match expected.
package stubutils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
)

// ReadFile is a wrapper function for ioutil.ReadFile.
var ReadFile = func(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// WriteFile is a wrapper function for ioutil.WriteFile.
var WriteFile = func(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

// ExecCommand returns output of named program with the given arguments.
var ExecCommand = func(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}

// List of mocked errors for testing
var (
	ErrOhNo = errors.New("oh no")
)

type (
	// Stubs is a wrapper for a stub command
	Stubs struct {
		queue           []ioStub
		t               tester
		origExecCommand func(string, ...string) ([]byte, error)
		origReadFile    func(string) ([]byte, error)
		origWriteFile   func(string, []byte, os.FileMode) error
	}

	ioStub interface {
		check(tester, []string) ([]byte, error)
	}

	// A tester is used by NewStubs.
	tester interface {
		Errorf(format string, args ...interface{})
		Fatalf(format string, args ...interface{})
	}

	// CmdStub implements ioStub interface and is used to mock a ExecCommand
	// function.
	CmdStub struct {
		// Expected program name.
		Cmd string
		// Expected  arguments.
		Args []string
		// Stub file path.
		StubFile string
		// Expected output or output from StubFile if provided.
		Output []byte
		// Expected error.
		Err error
	}

	// ReadFileStub implements ioStub interface and is used to mock a ReadFile
	// function.
	ReadFileStub struct {
		// Path to expected file.
		Path string
		// Path to stub file.
		StubFile string
		// Expected output or output from StubFile if provided.
		Output []byte
		// Expected error.
		Err error
	}

	// WriteFileStub implements ioStub interface and is used to mock a WriteFile
	// function.
	WriteFileStub struct {
		// Path to expected file.
		Path string
		// Writable data.
		Data []byte
		Mode os.FileMode
		// Expected error.
		Err error
	}
)

func (c *CmdStub) check(t tester, in []string) ([]byte, error) {
	// check command
	if c.Cmd != "" && c.Cmd != in[0] {
		t.Errorf("Expected command \"%s\"; received \"%s\"", c.Cmd, in[0])
	}

	// check arguments
	if len(c.Args) > 0 {
		if len(in) <= 1 {
			t.Errorf("Expected some arguments %+v; got none", c.Args)
		} else if !reflect.DeepEqual(c.Args, in[1:]) {
			t.Errorf("Expected arguments \"%+v\"; received \"%+v\"", c.Args, in[1:])
		}
	}

	// open the stub file if specified
	if c.StubFile != "" {
		b, err := ioutil.ReadFile(c.StubFile)
		if err != nil {
			t.Fatalf("Failed to read stub file %s; err = %v", c.StubFile, err)
		}
		c.Output = b
	}

	// return stubbed data
	return c.Output, c.Err
}

func (r *ReadFileStub) check(t tester, in []string) ([]byte, error) {
	path := in[0]

	// check requested path
	if r.Path != "" && r.Path != path {
		t.Errorf("Expected readFile's requested path to match \"%s\"; got \"%s\"", r.Path, path)
	}

	// open the stub file if specified
	if r.StubFile != "" {
		b, err := ioutil.ReadFile(r.StubFile)
		if err != nil {
			t.Fatalf("Failed to read stub file %s; err = %v", r.StubFile, err)
		}
		r.Output = b
	}

	// return stubbed response
	return r.Output, r.Err
}

func (w *WriteFileStub) check(t tester, in []string) ([]byte, error) {
	// check path
	if w.Path != "" && w.Path != in[0] {
		t.Errorf("Expected path to equal \"%s\"; got \"%s\"", w.Path, in[0])
	}

	// check contents
	if len(w.Data) > 0 && !reflect.DeepEqual(w.Data, []byte(in[1])) {
		t.Errorf("Excpected data to equal %v; got %v", string(w.Data), in[1])
	}

	// check mode
	if w.Mode != 0 && string(w.Mode) != in[2] {
		t.Errorf("Expected mode %v", w.Mode)
	}

	return nil, w.Err
}

// NewStubs creates a new stub collection and stubs the commands.
func NewStubs(t tester, stubs ...ioStub) *Stubs {
	s := &Stubs{
		t:               t,
		queue:           stubs,
		origExecCommand: ExecCommand,
		origReadFile:    ReadFile,
		origWriteFile:   WriteFile,
	}

	ReadFile = func(path string) ([]byte, error) {
		stub := s.get()

		// did we get something
		if stub == nil {
			return nil, fmt.Errorf("Stub queue is already empty when calling readFile(\"%s\")", path)
		}

		// validate the type
		if _, ok := stub.(*ReadFileStub); !ok {
			return nil, fmt.Errorf("Expected a readFile stub; got %T", stub)
		}

		// check it
		return stub.check(t, []string{path})
	}

	ExecCommand = func(cmd string, args ...string) ([]byte, error) {
		stub := s.get()

		// did we get something
		if stub == nil {
			return nil, fmt.Errorf("Stub queue is already empty when calling execCommand(\"%s\", %+v)", cmd, args)
		}

		// validate the type
		if _, ok := stub.(*CmdStub); !ok {
			return nil, fmt.Errorf("Expected a execCommand stub; got %T", stub)
		}

		// check it
		return stub.check(t, append([]string{cmd}, args...))
	}

	WriteFile = func(filename string, data []byte, perm os.FileMode) error {
		stub := s.get()

		// did we get something
		if stub == nil {
			return fmt.Errorf("Stub queue is already empty when calling writeFile(\"%s\" ...)", filename)
		}

		// validate the type
		if _, ok := stub.(*WriteFileStub); !ok {
			return fmt.Errorf("Expected a writeFile stub; got %T", stub)
		}

		// check it
		_, err := stub.check(t, []string{filename, string(data), string(perm)})
		return err
	}

	return s
}

// Add stubs to end of the queue
func (s *Stubs) Add(stub ...ioStub) {
	s.queue = append(s.queue, stub...)
}

// Get pops the first item from the queue. Returns nil if the queue is already empty.
func (s *Stubs) get() ioStub {
	if len(s.queue) == 0 {
		return nil
	}

	first := s.queue[0]
	s.queue = s.queue[1:]
	return first
}

// Close resets stubbed commands and and checks if the queue is empty.
func (s *Stubs) Close() {
	// reset functions
	ExecCommand = s.origExecCommand
	ReadFile = s.origReadFile
	WriteFile = s.origWriteFile

	if len(s.queue) > 0 {
		s.t.Errorf("Excpected the queue to be empty. %d items left in the queue", len(s.queue))
		for _, q := range s.queue {
			s.t.Errorf("Left in queue: %+v", q)
		}
	}
}
