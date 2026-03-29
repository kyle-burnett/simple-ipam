package delete

import (
	"os"
	"testing"

	"github.com/kyle-burnett/simple-ipam/internal/utils/testutils"
)

func Test_Delete(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testDelete.yaml")
	if err != nil {
		t.Fatalf("unexpected error creating test file: %v", err)
	}
	defer os.Remove(testFile)

	if err = Delete(testFile, "10.10.0.0/24", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want, err := os.ReadFile("testdata/delete_expected.yaml")
	if err != nil {
		t.Fatalf("unexpected error reading fixture: %v", err)
	}

	got, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func Test_DeleteRecursive(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testDeleteRecursive.yaml")
	if err != nil {
		t.Fatalf("unexpected error creating test file: %v", err)
	}
	defer os.Remove(testFile)

	if err = Delete(testFile, "10.10.0.0/20", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want, err := os.ReadFile("testdata/delete_recursive_expected.yaml")
	if err != nil {
		t.Fatalf("unexpected error reading fixture: %v", err)
	}

	got, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func Test_DeleteNoRecursive(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testDeleteNoRecursive.yaml")
	if err != nil {
		t.Fatalf("unexpected error creating test file: %v", err)
	}
	defer os.Remove(testFile)

	wantErr := "cannot delete 10.10.0.0/20 as subnets are defined under it. Use '-r' or '--recursive' to delete 10.10.0.0/20 and everything defined under it"
	err = Delete(testFile, "10.10.0.0/20", false)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != wantErr {
		t.Errorf("got error %q, want %q", err.Error(), wantErr)
	}
}
