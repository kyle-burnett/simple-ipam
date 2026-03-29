package add

import (
	"os"
	"testing"

	"github.com/kyle-burnett/simple-ipam/internal/utils/testutils"
)

func Test_AddSubnet(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testAdd.yaml")
	if err != nil {
		t.Fatalf("unexpected error creating test file: %v", err)
	}
	defer os.Remove(testFile)

	if err = Add(testFile, "10.10.0.0/25", "test subnet", []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want, err := os.ReadFile("testdata/add_subnet_expected.yaml")
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

func Test_AddSupernet(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testSupernet.yaml")
	if err != nil {
		t.Fatalf("unexpected error creating test file: %v", err)
	}
	defer os.Remove(testFile)

	if err = Add(testFile, "10.10.0.0/22", "test subnet", []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want, err := os.ReadFile("testdata/add_supernet_expected.yaml")
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

func Test_AddErrors(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testAddErrors.yaml")
	if err != nil {
		t.Fatalf("unexpected error creating test file: %v", err)
	}
	defer os.Remove(testFile)

	tests := []struct {
		name    string
		subnet  string
		wantErr string
	}{
		{
			name:    "invalid subnet",
			subnet:  "10.10.0.0/222",
			wantErr: "invalid subnet: error parsing existing CIDR: invalid CIDR address: 10.10.0.0/222",
		},
		{
			name:    "invalid notation",
			subnet:  "10.10.0.100/22",
			wantErr: "invalid subnet: 10.10.0.100/22 is not valid CIDR notation",
		},
		{
			name:    "duplicate subnet",
			subnet:  "10.10.0.0/20",
			wantErr: "error adding subnet: \"10.10.0.0/20\" already exists in this IPAM file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Add(testFile, tt.subnet, "test subnet", []string{})
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("got error %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
