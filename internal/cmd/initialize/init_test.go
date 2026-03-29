package initialize

import (
	"os"
	"testing"
)

func Test_InitCommand(t *testing.T) {
	if err := Initialize("test", "test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove("test.yaml") })

	want, err := os.ReadFile("testdata/init_expected.yaml")
	if err != nil {
		t.Fatalf("unexpected error reading fixture: %v", err)
	}

	got, err := os.ReadFile("test.yaml")
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func Test_InitCommand_FileAlreadyExists(t *testing.T) {
	if err := Initialize("test", "test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove("test.yaml") })

	wantErr := "IPAM file test.yaml already exists"
	err := Initialize("test", "test")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != wantErr {
		t.Errorf("got error %q, want %q", err.Error(), wantErr)
	}
}
