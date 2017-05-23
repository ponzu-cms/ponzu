package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrades your project to the current ponzu version",
	Long: `Will backup your own custom project code (like content, addons, uploads, etc)
if necessary. Before running '$ ponzu upgrade', you should update the 'ponzu'
package by running '$ go get -u github.com/ponzu-cms/ponzu/...'`,
	Example: `$ ponzu upgrade`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// confirm since upgrade will replace Ponzu core files
		path, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to find current directory: %s", err)
		}

		fmt.Println("Only files you added to this directory, 'addons' and 'content' will be preserved.")
		fmt.Println("Changes you made to Ponzu's internal code will be overwritten.")
		fmt.Println("Upgrade this project? (y/N):")

		answer, err := getAnswer()
		if err != nil {
			return err
		}

		switch answer {
		case "n", "no", "\r\n", "\n", "":
			fmt.Println("")

		case "y", "yes":
			err := upgradePonzuProjectDir(path)
			if err != nil {
				return err
			}

		default:
			fmt.Println("Input not recognized. No upgrade made. Answer as 'y' or 'n' only.")
		}
		return nil
	},
}

func upgradePonzuProjectDir(path string) error {
	core := []string{
		".gitattributes",
		"LICENSE",
		"ponzu-banner.png",
		"README.md",
		"cmd",
		"deployment",
		"management",
		"system",
	}

	stamp := fmt.Sprintf("ponzu-%d.bak", time.Now().Unix())
	temp := filepath.Join(os.TempDir(), stamp)
	err := os.Mkdir(temp, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	// track non-Ponzu core items (added by user)
	var user []os.FileInfo
	list, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, item := range list {
		// check if in core
		var isCore bool
		for _, name := range core {
			if item.Name() == name {
				isCore = true
				break
			}
		}

		if !isCore {
			user = append(user, item)
		}
	}

	// move non-Ponzu files to temp location
	fmt.Println("Preserving files to be restored after upgrade...")
	for _, item := range user {
		src := filepath.Join(path, item.Name())
		if item.IsDir() {
			err := os.Mkdir(filepath.Join(temp, item.Name()), os.ModeDir|os.ModePerm)
			if err != nil {
				return err
			}
		}

		err := copyAll(src, temp)
		if err != nil {
			return err
		}

		fmt.Println(" [-]", item.Name())

	}

	// remove all files in path
	for _, item := range list {
		err := os.RemoveAll(filepath.Join(path, item.Name()))
		if err != nil {
			return fmt.Errorf("Failed to remove old Ponzu files.\n%s", err)
		}
	}

	err = createProjectInDir(path)
	if err != nil {
		fmt.Println("")
		fmt.Println("Upgrade failed...")
		fmt.Println("Your code is backed up at the following location:")
		fmt.Println(temp)
		fmt.Println("")
		fmt.Println("Manually create a new Ponzu project here and copy those files within it to fully restore.")
		fmt.Println("")
		return err
	}

	// move non-Ponzu files from temp location backed
	restore, err := ioutil.ReadDir(temp)
	if err != nil {
		return err
	}

	fmt.Println("Restoring files preserved before upgrade...")
	for _, r := range restore {
		p := filepath.Join(temp, r.Name())
		err = copyAll(p, path)
		if err != nil {
			fmt.Println("Couldn't merge your previous project files with upgraded one.")
			fmt.Println("Manually copy your files from the following directory:")
			fmt.Println(temp)
			return err
		}

		fmt.Println(" [+]", r.Name())
	}

	// clean-up
	backups := []string{filepath.Join(path, stamp), temp}
	for _, bak := range backups {
		err := os.RemoveAll(bak)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	RegisterCmdlineCommand(upgradeCmd)
}
