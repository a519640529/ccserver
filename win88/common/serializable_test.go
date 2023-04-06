package common

import "testing"

func TestWriteMDumpFile(t *testing.T) {
	err := WriteMDumpFile("test.mdump", []byte("hello world!"))
	if err != nil {
		t.Fatal("WriteMDumpFile fail", err)
	}
}
