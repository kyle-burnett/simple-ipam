package initialize

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/fileutil"
)

var file, description string

var InitCmd = &cobra.Command{
	Use:          "init",
	Short:        "Initialize an empty IPAM file",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return Initialize(file, description)
	},
}

func init() {
	InitCmd.Flags().StringVarP(&file, "file", "f", "ipam", "Root IPAM file to create")
	InitCmd.Flags().StringVarP(&description, "description", "d", "", "Root IPAM file description")
}

func Initialize(file, description string) error {
	fileName := file + ".yaml"
	if _, err := os.Stat(fileName); err == nil {
		return fmt.Errorf("IPAM file %v already exists", fileName)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	ipam := models.IPAM{
		Subnets:     make(map[string]models.Subnets),
		Description: description,
	}

	return fileutil.WriteYAMLAtomic(fileName, &ipam)
}
