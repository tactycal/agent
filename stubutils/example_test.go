package stubutils

import "testing"

func ExampleNewStubs() {
	t := &testing.T{}
	s := NewStubs(t, &CmdStub{Cmd: "command", Args: []string{"-c"}, Output: []byte("out")})
	defer s.Close()

	if out, _ := ExecCommand("command", "-c"); string(out) != "out" {
		t.Errorf("Expected \"out\"; got %s", out)
	}
}

func ExampleCmdStub() {
	t := &testing.T{}
	s := NewStubs(t, &CmdStub{Cmd: "command", Args: []string{"-c"}, Output: []byte("out")})
	defer s.Close()

	if out, _ := ExecCommand("command", "-c"); string(out) != "out" {
		t.Errorf("Expected \"out\"; got %s", out)
	}
}

func ExampleReadFileStub() {
	t := &testing.T{}
	s := NewStubs(t, &ReadFileStub{Path: "/path", Output: []byte("out")})
	defer s.Close()

	if out, _ := ReadFile("/path"); string(out) != "out" {
		t.Errorf("Expected \"out\"; got %s", out)
	}
}

func ExampleWriteFileStub() {
	t := &testing.T{}
	s := NewStubs(t, &WriteFileStub{Path: "/path", Mode: 0600})
	defer s.Close()

	if err := WriteFile("/path", nil, 0600); err != nil {
		t.Errorf("Error not expected; got %s", err)
	}
}
