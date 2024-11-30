package testutil

import (
	"runtime"
	"strings"
	"testing"
)

func Must(t *testing.T, err error) {
	if err == nil {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("must: runtime call failed")
	}
	lastSlashIdx := strings.LastIndex(file, "/")
	if lastSlashIdx >= 0 && (len(file)-1) > lastSlashIdx {
		file = file[lastSlashIdx+1:]
	}
	t.Fatalf("\r%s:%d: %s\n", file, line, err.Error())
}

func MustIdx(t *testing.T, idx int, err error) {
	if err == nil {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("must: runtime call failed")
	}
	lastSlashIdx := strings.LastIndex(file, "/")
	if lastSlashIdx >= 0 && (len(file)-1) > lastSlashIdx {
		file = file[lastSlashIdx+1:]
	}
	t.Fatalf("\r%s:%d: idx %d: %s\n", file, line, idx, err.Error())
}
