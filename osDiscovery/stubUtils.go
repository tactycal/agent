package osDiscovery

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

var readFile = func(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

var writeFile = func(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

var execCommand = func(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}

var ohNoErr = errors.New("oh no")

type (
	stub struct {
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
		Cmd      string
		Args     []string
		StubFile string
		Output   []byte
		Err      error
	}

	readFileStub struct {
		Path     string
		StubFile string
		Output   []byte
		Err      error
	}

	writeFileStub struct {
		Path string
		Data []byte
		Mode os.FileMode
		Err  error
	}
)

func (c *cmdStub) check(t *testing.T, in []string) ([]byte, error) {
	// check number of parameters
	if len(in) < 1 {
		t.Fatalf("Command not defined")
	}

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

func (r *readFileStub) check(t *testing.T, in []string) ([]byte, error) {
	if len(in) != 1 {
		t.Fatal("Expected one argument to be passed to readFile; got none")
	}

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

func (w *writeFileStub) check(t *testing.T, in []string) ([]byte, error) {
	// check path
	if w.Path != "" && w.Path != in[0] {
		t.Errorf("Expected path to equal \"%s\"; got \"%s\"", w.Path, in[0])
	}

	// check contents
	if len(w.Data) > 0 && !reflect.DeepEqual(w.Data, []byte(in[1])) {
		t.Errorf("Excpected data to equal %v; got %v", w.Data, in[1])
	}

	// check mode
	if w.Mode != 0 && string(w.Mode) != in[2] {
		t.Errorf("Expected mode %v", w.Mode)
	}

	return nil, w.Err
}

// Creates a new stub collection and stubs the commands.
func newStubs(t *testing.T, stubs ...ioStub) *stub {
	s := &stub{
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

// Add a single stub to end of the queue.
func (s *stub) Add(stub ...ioStub) {
	s.queue = append(s.queue, stub...)
}

// Get pops the first item from the queue. Returns nil if the queue is already
// empty.
func (s *stub) Get() ioStub {
	if len(s.queue) == 0 {
		return nil
	}

	first := s.queue[0]
	s.queue = s.queue[1:]
	return first
}

// Resets stubbed commands and checks if the queue is empty.
func (s *stub) Close() {
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
