package addnextavailable

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/fileutil"
	"github.com/kyle-burnett/simple-ipam/internal/utils/subnetutils"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

var parent, description, inputFile string
var subnetToAdd int
var tags []string

var AddNextAvailableCmd = &cobra.Command{
	Use:          "add-next-available",
	Short:        "Add the next available subnet of a given length under a parent subnet",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return AddNextAvailable(inputFile, parent, description, subnetToAdd, tags)
	},
}

func init() {
	AddNextAvailableCmd.Flags().IntVarP(&subnetToAdd, "prefix-length", "l", 0, "prefix length (CIDR mask bits) of the subnet to allocate")
	AddNextAvailableCmd.Flags().StringVarP(&parent, "parent", "p", "", "Parent subnet")
	AddNextAvailableCmd.Flags().StringVarP(&inputFile, "file", "f", "", "ipam file")
	_ = AddNextAvailableCmd.MarkFlagRequired("prefix-length")
	_ = AddNextAvailableCmd.MarkFlagRequired("file")
	_ = AddNextAvailableCmd.MarkFlagRequired("parent")
	AddNextAvailableCmd.Flags().StringVarP(&description, "description", "d", "", "description for the subnet")
	AddNextAvailableCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags to add to the subnet")
}

func AddNextAvailable(inputFile, parent, description string, subnetToAdd int, tags []string) error {
	err := subnetutils.CheckValidSubnet(parent)
	if err != nil {
		return err
	}

	if subnetToAdd < 1 || subnetToAdd > 32 {
		return fmt.Errorf("%v is not a valid IPv4 CIDR mask. Must be > 0 and <= 32", subnetToAdd)
	}

	ipamData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading IPAM file: %v", err)
	}

	var ipam models.IPAM
	err = yaml.Unmarshal(ipamData, &ipam)
	if err != nil {
		return fmt.Errorf("error unmarshaling IPAM: %v", err)
	}

	_, parentNet, err := net.ParseCIDR(parent)
	if err != nil {
		return err
	}

	err = withParent(ipam.Subnets, parent, func(p *models.Subnets) error {
		descendants, err := collectDescendants(p.Subnets)
		if err != nil {
			return err
		}
		chosen, err := findNextAvailable(parentNet, subnetToAdd, descendants)
		if err != nil {
			return err
		}
		return insertAtDeepest(p.Subnets, chosen, models.Subnets{
			Description: description,
			Tags:        tags,
			Subnets:     map[string]models.Subnets{},
		})
	})
	if err != nil {
		return err
	}

	return fileutil.WriteYAMLAtomic(inputFile, &ipam)
}

func withParent(allSubnets map[string]models.Subnets, parentCIDR string, fn func(parent *models.Subnets) error) error {
	for subnet, values := range allSubnets {
		if subnet == parentCIDR {
			if values.Subnets == nil {
				values.Subnets = make(map[string]models.Subnets)
			}
			if err := fn(&values); err != nil {
				return err
			}
			allSubnets[subnet] = values // write the copy back
			return nil
		}
		isSubnet, err := subnetutils.IsSubnetOf(subnet, parentCIDR)
		if err != nil {
			return err
		}

		if isSubnet {
			return withParent(values.Subnets, parentCIDR, fn)
		}
	}

	return fmt.Errorf("parent subnet %q does not exist in IPAM data", parentCIDR)
}

// collectDescendants walks the subtree rooted at tree and returns every
// subnet it contains, parsed to *net.IPNet. Errors on any malformed CIDR key.
func collectDescendants(tree map[string]models.Subnets) ([]*net.IPNet, error) {
	var out []*net.IPNet
	var walk func(m map[string]models.Subnets) error
	walk = func(m map[string]models.Subnets) error {
		for cidr, node := range m {
			_, n, err := net.ParseCIDR(cidr)
			if err != nil {
				return fmt.Errorf("corrupt IPAM: %q: %w", cidr, err)
			}
			out = append(out, n)
			if err := walk(node.Subnets); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(tree); err != nil {
		return nil, err
	}
	return out, nil
}

func findNextAvailable(parentNet *net.IPNet, subnetToAdd int, descendants []*net.IPNet) (*net.IPNet, error) {
	parentNetSize, bits := parentNet.Mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("only IPv4 is supported")
	}
	if parentNetSize == 0 {
		return nil, fmt.Errorf("/0 parents are not supported")
	}
	if parentNetSize >= subnetToAdd {
		return nil, fmt.Errorf("desired prefix /%d must be longer than parent /%d and <= 32", subnetToAdd, parentNetSize)
	}

	start := binary.BigEndian.Uint32(parentNet.IP.To4())
	blockSize := uint32(1) << uint(32-subnetToAdd)            // addresses per candidate
	numBlocks := uint32(1) << uint(subnetToAdd-parentNetSize) // candidates to try
	mask := net.CIDRMask(subnetToAdd, 32)

	for i := range numBlocks {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, start+i*blockSize)
		candidate := &net.IPNet{IP: ip, Mask: mask}

		if !candidateBlocked(candidate, subnetToAdd, descendants) {
			return candidate, nil
		}
	}
	return nil, fmt.Errorf("no available /%d subnet in %s", subnetToAdd, parentNet)
}

// candidateBlocked reports whether the candidate would displace existing
// address space. A descendant blocks the candidate iff its range is fully
// within (or equal to) the candidate's range — i.e., the descendant's
// prefix is at least as long as the candidate's and its network IP falls
// inside the candidate. Descendants that are strict supernets of the
// candidate are not blockers; they are containers the candidate can nest
// inside.
func candidateBlocked(candidate *net.IPNet, candOnes int, descendants []*net.IPNet) bool {
	for _, d := range descendants {
		dOnes, _ := d.Mask.Size()
		if dOnes < candOnes {
			continue // d is bigger than candidate; it is a potential container, not a blocker
		}
		if candidate.Contains(d.IP) {
			return true
		}
	}
	return false
}

// insertAtDeepest inserts entry under candidate's deepest existing ancestor
// in tree. tree is the direct-children map of the parent; candidate is
// assumed to be free (not blocked by any descendant) and strictly inside
// the parent's range. The CIDR invariant (at most one sibling at a given
// level can contain a given address) guarantees only one branch matches.
func insertAtDeepest(tree map[string]models.Subnets, candidate *net.IPNet, entry models.Subnets) error {
	candOnes, _ := candidate.Mask.Size()

	for cidr, values := range tree {
		_, existing, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("corrupt IPAM: %q: %w", cidr, err)
		}
		existingOnes, _ := existing.Mask.Size()

		if existingOnes < candOnes && existing.Contains(candidate.IP) {
			if values.Subnets == nil {
				values.Subnets = make(map[string]models.Subnets)
			}
			if err := insertAtDeepest(values.Subnets, candidate, entry); err != nil {
				return err
			}
			tree[cidr] = values // write-back for nil-init propagation
			return nil
		}
	}

	tree[candidate.String()] = entry
	return nil
}
