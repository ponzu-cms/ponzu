package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func newProjectInDir(path string) error {
	// set path to be nested inside $GOPATH/src
	gopath := os.Getenv("GOPATH")
	path = filepath.Join(gopath, "src", path)

	// check if anything exists at the path, ask if it should be overwritten
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		fmt.Println("Path exists, overwrite contents? (y/N):")

		var answer string
		_, err := fmt.Scanf("%s\n", &answer)
		if err != nil {
			if err.Error() == "unexpected newline" {
				answer = ""
			} else {
				return err
			}
		}

		answer = strings.ToLower(answer)

		switch answer {
		case "n", "no", "\r\n", "\n", "":
			fmt.Println("")

		case "y", "yes":
			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("Failed to overwrite %s. \n%s", path, err)
			}

			return createProjInDir(path)

		default:
			fmt.Println("Input not recognized. No files overwritten. Answer as 'y' or 'n' only.")
		}

		return nil
	}

	return createProjInDir(path)
}

var ponzuRepo = []string{"github.com", "bosssauce", "ponzu"}

func createProjInDir(path string) error {
	gopath := os.Getenv("GOPATH")
	repo := ponzuRepo
	local := filepath.Join(gopath, "src", filepath.Join(repo...))
	network := "https://" + strings.Join(repo, "/") + ".git"

	// create the directory or overwrite it
	err := os.MkdirAll(path, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	if dev {
		if fork != "" {
			local = filepath.Join(gopath, "src", fork)
		}

		devClone := exec.Command("git", "clone", local, "--branch", "ponzu-dev", "--single-branch", path)
		devClone.Stdout = os.Stdout
		devClone.Stderr = os.Stderr

		err = devClone.Start()
		if err != nil {
			return err
		}

		err = devClone.Wait()
		if err != nil {
			return err
		}

		err = vendorCorePackages(path)
		if err != nil {
			return err
		}

		fmt.Println("Dev build cloned from " + local + ":ponzu-dev")
		return nil
	}

	// try to git clone the repository from the local machine's $GOPATH
	localClone := exec.Command("git", "clone", local, path)
	localClone.Stdout = os.Stdout
	localClone.Stderr = os.Stderr

	err = localClone.Start()
	if err != nil {
		return err
	}
	err = localClone.Wait()
	if err != nil {
		fmt.Println("Couldn't clone from", local, ". Trying network...")

		// try to git clone the repository over the network
		networkClone := exec.Command("git", "clone", network, path)
		networkClone.Stdout = os.Stdout
		networkClone.Stderr = os.Stderr

		err = networkClone.Start()
		if err != nil {
			fmt.Println("Network clone failed to start. Try again and make sure you have a network connection.")
			return err
		}
		err = networkClone.Wait()
		if err != nil {
			fmt.Println("Network clone failure.")
			// failed
			return fmt.Errorf("Failed to clone files from local machine [%s] and over the network [%s].\n%s", local, network, err)
		}
	}

	// create an internal vendor directory in ./cmd/ponzu and move content,
	// management and system packages into it
	err = vendorCorePackages(path)
	if err != nil {
		return err
	}

	gitDir := filepath.Join(path, ".git")
	err = os.RemoveAll(gitDir)
	if err != nil {
		fmt.Println("Failed to remove .git directory from your project path. Consider removing it manually.")
	}

	fmt.Println("New ponzu project created at", path)
	return nil
}

func vendorCorePackages(path string) error {
	vendorPath := filepath.Join(path, "cmd", "ponzu", "vendor", "github.com", "bosssauce", "ponzu")
	err := os.MkdirAll(vendorPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	// create a user content directory to be vendored
	contentPath := filepath.Join(path, "content")
	err = os.Mkdir(contentPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	dirs := []string{"content", "management", "system"}
	for _, dir := range dirs {
		err = os.Rename(filepath.Join(path, dir), filepath.Join(vendorPath, dir))
		if err != nil {
			return err
		}
	}

	// create a user content directory at project root
	contentPath = filepath.Join(path, "content")
	err = os.Mkdir(contentPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	noRoot := strings.Split(src, string(filepath.Separator))[1:]
	path := filepath.Join(noRoot...)
	dstFile, err := os.Create(filepath.Join(dst, path))
	defer dstFile.Close()
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	defer srcFile.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func copyFilesWarnConflicts(srcDir, dstDir string, conflicts []string) error {
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if len(path) > len(srcDir) {
				path = path[len(srcDir)+1:]
			}
			dir := filepath.Join(dstDir, path)
			err := os.MkdirAll(dir, os.ModeDir|os.ModePerm)
			if err != nil {
				return err
			}

			return nil
		}

		for _, conflict := range conflicts {
			if info.Name() == conflict {
				fmt.Println("Ponzu couldn't fully build your project:")
				fmt.Println("You must rename the following file, as it conflicts with Ponzu core:")
				fmt.Println(path)
				fmt.Println("")
				fmt.Println("Once the files above have been renamed, run '$ ponzu build' to retry.")
				return errors.New("Ponzu has very few internal conflicts, sorry for the inconvenience.")
			}
		}

		err = copyFile(path, dstDir)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func buildPonzuServer(args []string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// copy all ./content files to internal vendor directory
	src := "content"
	dst := filepath.Join("cmd", "ponzu", "vendor", "github.com", "bosssauce", "ponzu", "content")
	err = copyFilesWarnConflicts(src, dst, nil)
	if err != nil {
		return err
	}

	// copy all ./addons files & dirs to internal vendor directory
	src = "addons"
	dst = filepath.Join("cmd", "ponzu", "vendor")
	err = copyFilesWarnConflicts(src, dst, nil)
	if err != nil {
		return err
	}

	// execute go build -o ponzu-cms cmd/ponzu/*.go
	mainPath := filepath.Join(pwd, "cmd", "ponzu", "main.go")
	optsPath := filepath.Join(pwd, "cmd", "ponzu", "options.go")
	genPath := filepath.Join(pwd, "cmd", "ponzu", "generate.go")
	build := exec.Command("go", "build", "-o", "ponzu-server", mainPath, optsPath, genPath)
	build.Stderr = os.Stderr
	build.Stdout = os.Stdout

	err = build.Start()
	if err != nil {
		return errors.New("Ponzu build step failed. Please try again. " + "\n" + err.Error())

	}
	err = build.Wait()
	if err != nil {
		return errors.New("Ponzu build step failed. Please try again. " + "\n" + err.Error())

	}

	return nil
}
