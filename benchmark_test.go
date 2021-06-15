package execpool

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var testPool *Pool

const message = "this is nonesense"

func TestMain(m *testing.M) {
	var err error
	cmd := exec.Command("grep", "none")
	testPool, err = New(cmd, 100)
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func BenchmarkNew(b *testing.B) {
	// run the Fib function b.N times

	reader := strings.NewReader(message)

	for n := 0; n < b.N; n++ {
		stdout := testPool.Exec(reader)
		out, err := ioutil.ReadAll(stdout)
		if err != nil {
			b.Fatalf("failed to read stdout: %v", err)
		}

		if string(out) != "this is nonesense\n" {
			b.Fatalf("string doesn't match: %v", string(out))
		}

		reader.Reset(message)
	}
}

func BenchmarkCmd(b *testing.B) {
	// run the Fib function b.N times
	buf := bytes.NewBuffer([]byte{})
	const message = "this is nonesense"
	stdin := strings.NewReader(message)

	for n := 0; n < b.N; n++ {
		cmd := exec.Command("grep", "none")
		cmd.Stdin = stdin
		cmd.Stdout = buf

		if err := cmd.Run(); err != nil {
			b.Fatalf("cmd.Run failed: %v", err)
		}

		if buf.String() != "this is nonesense\n" {
			b.Fatalf("strings don't match: %s", buf.String())
		}
		buf.Reset()
		stdin.Reset(message)
	}
}
