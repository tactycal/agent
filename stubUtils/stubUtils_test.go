package stubUtils

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

type testerMock struct {
	expectedErrMsgs []string
	t               *testing.T
}

func (tM *testerMock) Errorf(format string, args ...interface{}) {
	errMsg := fmt.Sprintf(format, args...)
	tM.checkErrMsg(errMsg)
}

func (tM *testerMock) Fatalf(format string, args ...interface{}) {
	errMsg := fmt.Sprintf(format, args...)
	tM.checkErrMsg(errMsg)
}

func (tM *testerMock) checkErrMsg(errMsg string) {
	var expectedErrMsg string
	if len(tM.expectedErrMsgs) > 0 {
		expectedErrMsg = tM.expectedErrMsgs[0]
		tM.expectedErrMsgs = tM.expectedErrMsgs[1:]
	}

	if expectedErrMsg != errMsg {
		tM.t.Errorf("Expected error message\n%s\n doesn't match received error message\n%s\n", expectedErrMsg, errMsg)
	}
}

func TestCmdStub(t *testing.T) {
	tM := &testerMock{t: t}
	s := NewStubs(tM)
	defer s.Close()

	// 1. command doesn't match expected
	tM.expectedErrMsgs = append(tM.expectedErrMsgs, `Expected command "expectedCommand"; received "command"`)
	s.Add(&CmdStub{Cmd: "expectedCommand"})
	ExecCommand("command")

	// 2. arguments have been expected
	tM.expectedErrMsgs = append(tM.expectedErrMsgs, `Expected some arguments [arg]; got none`)
	s.Add(&CmdStub{Cmd: "command", Args: []string{"arg"}})
	ExecCommand("command")

	// 3. arguments don't match
	tM.expectedErrMsgs = append(tM.expectedErrMsgs, `Expected arguments "[arg]"; received "[gra]"`)
	s.Add(&CmdStub{Cmd: "command", Args: []string{"arg"}})
	ExecCommand("command", "gra")

	// 4. stub file doesn't exist
	tM.expectedErrMsgs = append(tM.expectedErrMsgs, `Failed to read stub file nofile; err = open nofile: no such file or directory`)
	s.Add(&CmdStub{Cmd: "command", StubFile: "nofile"})
	ExecCommand("command")

	// 5. check stub file content
	tmpFile, _ := ioutil.TempFile("", "tmpFile")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("tmpFile"))
	s.Add(&CmdStub{Cmd: "command", StubFile: tmpFile.Name()})
	b, _ := ExecCommand("command")
	if !reflect.DeepEqual(string(b), "tmpFile") {
		t.Errorf("File content\n%s\n doesn't match expected\n%s\n", string(b), "tmpFile")
	}

	// 6. empty stub queue
	expectedErr := fmt.Errorf(`Stub queue is already empty when calling execCommand("command", [arg])`)
	_, err := ExecCommand("command", "arg")
	if !reflect.DeepEqual(err, expectedErr) {
		t.Errorf("Expected error\n%s\ngot\n%s\n", expectedErr, err)
	}

	// 7. wrong stub type
	expectedErr = fmt.Errorf(`Expected a execCommand stub; got *stubUtils.ReadFileStub`)
	s.Add(&ReadFileStub{Path: "/path"})
	_, err = ExecCommand("command", "arg")
	if !reflect.DeepEqual(err, expectedErr) {
		t.Errorf("Expected error\n%s\ngot\n%s\n", expectedErr, err)
	}

	// 8. check that no error message left
	tM.checkErrMsg("")
}

func TestReadFileStub(t *testing.T) {
	tM := &testerMock{t: t}
	s := NewStubs(tM)
	defer s.Close()

	// 1. check expected path
	tM.expectedErrMsgs = append(tM.expectedErrMsgs, `Expected readFile's requested path to match "/path"; got "/nopath"`)
	s.Add(&ReadFileStub{Path: "/path"})
	ReadFile("/nopath")

	// 2. stub file doesn't exist
	tM.expectedErrMsgs = append(tM.expectedErrMsgs, `Failed to read stub file /nofile; err = open /nofile: no such file or directory`)
	s.Add(&ReadFileStub{Path: "/path", StubFile: "/nofile"})
	ReadFile("/path")

	// 3. check stub file content
	tmpFile, _ := ioutil.TempFile("", "tmpFile")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("tmpFile"))
	s.Add(&ReadFileStub{Path: "/path", StubFile: tmpFile.Name()})

	b, _ := ReadFile("/path")
	if !reflect.DeepEqual(string(b), "tmpFile") {
		t.Errorf("File content\n%s\n doesn't match expected\n%s\n", string(b), "tmpFile")
	}

	// 4. empty stub queue
	expectedErr := fmt.Errorf(`Stub queue is already empty when calling readFile("/path")`)

	_, err := ReadFile("/path")
	if !reflect.DeepEqual(err, expectedErr) {
		t.Errorf("Expected error\n%s\ngot\n%s\n", expectedErr, err)
	}

	// 5. wrong stub type
	expectedErr = fmt.Errorf(`Expected a readFile stub; got *stubUtils.CmdStub`)
	s.Add(&CmdStub{Cmd: "command"})

	_, err = ReadFile("/path")
	if !reflect.DeepEqual(err, expectedErr) {
		t.Errorf("Expected error\n%s\ngot\n%s\n", expectedErr, err)
	}

	// 6. check that no error message left
	tM.checkErrMsg("")

}

func TestWriteFileStub(t *testing.T) {
	tM := &testerMock{t: t}
	s := NewStubs(tM)
	defer s.Close()

	// 1. check expected path
	tM.expectedErrMsgs = append(tM.expectedErrMsgs, `Expected path to equal "/path"; got "/nopath"`)
	s.Add(&WriteFileStub{Path: "/path"})
	WriteFile("/nopath", nil, 0600)

	// 2. writable data doesn't match
	tM.expectedErrMsgs = append(tM.expectedErrMsgs, `Excpected data to equal match; got no match`)
	s.Add(&WriteFileStub{Path: "/path", Data: []byte("match")})
	WriteFile("/path", []byte("no match"), 0600)

	// 3. check mode
	tM.expectedErrMsgs = append(tM.expectedErrMsgs, `Expected mode -rwx------`)
	s.Add(&WriteFileStub{Path: "/path", Mode: 0700})
	WriteFile("/path", nil, 0600)

	// 4. empty stub queue
	expectedErr := fmt.Errorf(`Stub queue is already empty when calling writeFile("/path" ...)`)

	err := WriteFile("/path", nil, 0600)
	if !reflect.DeepEqual(err, expectedErr) {
		t.Errorf("Expected error\n%s\ngot\n%s\n", expectedErr, err)
	}

	// 5. wrong stub type
	expectedErr = fmt.Errorf(`Expected a writeFile stub; got *stubUtils.CmdStub`)
	s.Add(&CmdStub{Cmd: "command"})

	err = WriteFile("/path", nil, 0600)
	if !reflect.DeepEqual(err, expectedErr) {
		t.Errorf("Expected error\n%s\ngot\n%s\n", expectedErr, err)
	}

	// 6. check that no error message left
	tM.checkErrMsg("")
}

func TestStubQueue(t *testing.T) {
	tM := &testerMock{t: t}
	s := NewStubs(tM)

	// 1. queue is not empty
	s.Add(&CmdStub{Cmd: "command"})

	tM.expectedErrMsgs = append(tM.expectedErrMsgs,
		`Excpected the queue to be empty. 1 items left in the queue`,
		`Left in queue: &{Cmd:command Args:[] StubFile: Output:[] Err:<nil>}`)
	s.Close()

	// 2. check that no error message left
	tM.checkErrMsg("")

}
