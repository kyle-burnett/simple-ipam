package addnextavailable

import (
	"os"
	"testing"

	"github.com/kyle-burnett/simple-ipam/internal/utils/testutils"
)

// writeSeedFile writes the given YAML content to fileName and registers
// cleanup. Used by tests that need a starting state other than the
// default produced by testutils.CreateTestFile.
func writeSeedFile(t *testing.T, fileName, content string) string {
	t.Helper()
	if err := os.WriteFile(fileName, []byte(content), 0o644); err != nil {
		t.Fatalf("unexpected error writing seed file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(fileName) })
	return fileName
}

// assertGolden compares the contents of gotPath to the golden file at
// goldenPath byte-for-byte, matching the style used by the `add` package tests.
func assertGolden(t *testing.T, gotPath, goldenPath string) {
	t.Helper()
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("unexpected error reading output: %v", err)
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("unexpected error reading fixture: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// Happy path: empty /24 parent, request a /26. Expect 10.10.0.0/26 inserted
// as a direct child of 10.10.0.0/24 under the testutils default seed.
func Test_AddNextAvailable_BasicAllocation(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testBasic.yaml")
	if err != nil {
		t.Fatalf("unexpected error creating test file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(testFile) })

	if err := AddNextAvailable(testFile, "10.10.0.0/24", "first /26", 26, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, testFile, "testdata/basic_allocation_expected.yaml")
}

// Lowest-address-first: a pre-existing child in the middle of the parent
// must not prevent the allocator from reusing the hole before it.
func Test_AddNextAvailable_HoleReuse(t *testing.T) {
	seed := `description: ""
subnets:
    10.0.0.0/24:
        description: parent
        tags: []
        subnets:
            10.0.0.64/26:
                description: pre-existing
                tags: []
                subnets: {}
`
	testFile := writeSeedFile(t, "testHole.yaml", seed)

	if err := AddNextAvailable(testFile, "10.0.0.0/24", "hole", 26, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, testFile, "testdata/hole_reuse_expected.yaml")
}

// Four consecutive allocations under a /24 must fill all four /26 slots
// in ascending order.
func Test_AddNextAvailable_FillAllSlots(t *testing.T) {
	seed := `description: ""
subnets:
    10.0.0.0/24:
        description: parent
        tags: []
        subnets: {}
`
	testFile := writeSeedFile(t, "testFill.yaml", seed)

	for i, desc := range []string{"slot 1", "slot 2", "slot 3", "slot 4"} {
		if err := AddNextAvailable(testFile, "10.0.0.0/24", desc, 26, []string{}); err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
	}
	assertGolden(t, testFile, "testdata/fill_all_slots_expected.yaml")
}

// Mixed sizes: a /26 exists; asking for a /25 must skip 10.0.0.0/25
// (which contains the existing /26) and return 10.0.0.128/25.
func Test_AddNextAvailable_MixedSizeOverlap(t *testing.T) {
	seed := `description: ""
subnets:
    10.0.0.0/24:
        description: parent
        tags: []
        subnets:
            10.0.0.64/26:
                description: existing /26
                tags: []
                subnets: {}
`
	testFile := writeSeedFile(t, "testMixed.yaml", seed)

	if err := AddNextAvailable(testFile, "10.0.0.0/24", "upper half", 25, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, testFile, "testdata/mixed_size_overlap_expected.yaml")
}

// Recursive parent lookup: the parent /24 is nested three levels deep.
// findParent must descend through /16 and /20 to locate it.
func Test_AddNextAvailable_NestedParent(t *testing.T) {
	seed := `description: ""
subnets:
    10.0.0.0/16:
        description: outer
        tags: []
        subnets:
            10.0.0.0/20:
                description: middle
                tags: []
                subnets:
                    10.0.0.0/24:
                        description: inner
                        tags: []
                        subnets: {}
`
	testFile := writeSeedFile(t, "testNested.yaml", seed)

	if err := AddNextAvailable(testFile, "10.0.0.0/24", "deep", 26, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, testFile, "testdata/nested_parent_expected.yaml")
}

// Deep descent: an empty /25 direct child of the parent /24 is a container,
// not a blocker. Asking for a /26 under the /24 must nest 10.0.0.0/26 inside
// the empty /25 rather than placing it as a sibling at the /24 level.
func Test_AddNextAvailable_DescendsIntoEmptyChild(t *testing.T) {
	seed := `description: ""
subnets:
    10.0.0.0/24:
        description: parent
        tags: []
        subnets:
            10.0.0.0/25:
                description: existing /25
                tags: []
                subnets: {}
`
	testFile := writeSeedFile(t, "testDescendEmpty.yaml", seed)

	if err := AddNextAvailable(testFile, "10.0.0.0/24", "nested /26", 26, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, testFile, "testdata/descends_into_empty_child_expected.yaml")
}

// Deep descent past a non-empty container: parent /24 > /25 > /26. Asking
// for another /26 under the /24 must find 10.0.0.64/26 (the remaining free
// half of the /25) and nest it alongside the existing /26 inside the /25 —
// not place it at the /24 level.
func Test_AddNextAvailable_DescendsPastGrandchild(t *testing.T) {
	seed := `description: ""
subnets:
    10.0.0.0/24:
        description: parent
        tags: []
        subnets:
            10.0.0.0/25:
                description: middle
                tags: []
                subnets:
                    10.0.0.0/26:
                        description: grandchild
                        tags: []
                        subnets: {}
`
	testFile := writeSeedFile(t, "testDescendPast.yaml", seed)

	if err := AddNextAvailable(testFile, "10.0.0.0/24", "new /26", 26, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, testFile, "testdata/descends_past_grandchild_expected.yaml")
}

// Edge prefix /31 under /30: allocate both available slots, confirm a
// third request fails with the exhaustion error.
func Test_AddNextAvailable_EdgePrefix31(t *testing.T) {
	seed := `description: ""
subnets:
    10.0.0.0/30:
        description: parent
        tags: []
        subnets: {}
`
	testFile := writeSeedFile(t, "testEdge31.yaml", seed)

	if err := AddNextAvailable(testFile, "10.0.0.0/30", "first", 31, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := AddNextAvailable(testFile, "10.0.0.0/30", "second", 31, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := AddNextAvailable(testFile, "10.0.0.0/30", "third", 31, []string{})
	if err == nil {
		t.Fatalf("expected exhaustion error on third allocation, got nil")
	}
	want := "no available /31 subnet in 10.0.0.0/30"
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
	assertGolden(t, testFile, "testdata/edge_prefix_31_expected.yaml")
}

// Edge prefix /32 under /30: allocate all four host addresses, confirm
// a fifth request fails with the exhaustion error.
func Test_AddNextAvailable_EdgePrefix32(t *testing.T) {
	seed := `description: ""
subnets:
    10.0.0.0/30:
        description: parent
        tags: []
        subnets: {}
`
	testFile := writeSeedFile(t, "testEdge32.yaml", seed)

	for i := 1; i <= 4; i++ {
		if err := AddNextAvailable(testFile, "10.0.0.0/30", "", 32, []string{}); err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
	}
	err := AddNextAvailable(testFile, "10.0.0.0/30", "", 32, []string{})
	if err == nil {
		t.Fatalf("expected exhaustion error on fifth allocation, got nil")
	}
	assertGolden(t, testFile, "testdata/edge_prefix_32_expected.yaml")
}

// Exhaustion: a /24 fully covered by four /26 children must reject a
// fifth /26 request with a clean error.
func Test_AddNextAvailable_Exhaustion(t *testing.T) {
	seed := `description: ""
subnets:
    10.0.0.0/24:
        description: parent
        tags: []
        subnets:
            10.0.0.0/26:
                description: ""
                tags: []
                subnets: {}
            10.0.0.64/26:
                description: ""
                tags: []
                subnets: {}
            10.0.0.128/26:
                description: ""
                tags: []
                subnets: {}
            10.0.0.192/26:
                description: ""
                tags: []
                subnets: {}
`
	testFile := writeSeedFile(t, "testExhaust.yaml", seed)

	err := AddNextAvailable(testFile, "10.0.0.0/24", "", 26, []string{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	want := "no available /26 subnet in 10.0.0.0/24"
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
}

// Table-driven error cases that share the testutils default seed
// (10.10.0.0/20 > 10.10.0.0/24).
func Test_AddNextAvailable_Errors(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testErrors.yaml")
	if err != nil {
		t.Fatalf("unexpected error creating test file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(testFile) })

	tests := []struct {
		name    string
		parent  string
		prefix  int
		wantErr string
	}{
		{
			name:    "parent not in IPAM",
			parent:  "192.168.0.0/24",
			prefix:  26,
			wantErr: `parent subnet "192.168.0.0/24" does not exist in IPAM data`,
		},
		{
			name:    "non-canonical parent",
			parent:  "10.10.0.5/24",
			prefix:  26,
			wantErr: "10.10.0.5/24 is not valid CIDR notation",
		},
		{
			name:    "desired prefix equal to parent",
			parent:  "10.10.0.0/24",
			prefix:  24,
			wantErr: "desired prefix /24 must be longer than parent /24 and <= 32",
		},
		{
			name:    "desired prefix shorter than parent",
			parent:  "10.10.0.0/24",
			prefix:  20,
			wantErr: "desired prefix /20 must be longer than parent /24 and <= 32",
		},
		{
			name:    "CIDR mask too long",
			parent:  "10.10.0.0/24",
			prefix:  33,
			wantErr: "33 is not a valid IPv4 CIDR mask. Must be > 0 and <= 32",
		},
		{
			name:    "CIDR mask too short",
			parent:  "10.10.0.0/24",
			prefix:  0,
			wantErr: "0 is not a valid IPv4 CIDR mask. Must be > 0 and <= 32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AddNextAvailable(testFile, tt.parent, "", tt.prefix, []string{})
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("got error %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
