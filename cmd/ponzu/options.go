package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func newProjectInDir(path string) error {
	// set path to be nested inside $GOPATH/src
	gopath := os.Getenv("GOPATH")
	path = filepath.Join(gopath, "src", path)

	// check if anything exists at the path, ask if it should be overwritten
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		fmt.Println("Path exists, overwrite contents? (y/N):")

		answer, err := getAnswer()
		if err != nil {
			return err
		}

		switch answer {
		case "n", "no", "\r\n", "\n", "":
			fmt.Println("")

		case "y", "yes":
			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("Failed to overwrite %s. \n%s", path, err)
			}

			return createProjectInDir(path)

		default:
			fmt.Println("Input not recognized. No files overwritten. Answer as 'y' or 'n' only.")
		}

		return nil
	}

	return createProjectInDir(path)
}

var ponzuRepo = []string{"github.com", "ponzu-cms", "ponzu"}

func getAnswer() (string, error) {
	var answer string
	_, err := fmt.Scanf("%s\n", &answer)
	if err != nil {
		if err.Error() == "unexpected newline" {
			answer = ""
		} else {
			return "", err
		}
	}

	answer = strings.ToLower(answer)

	return answer, nil
}

func createProjectInDir(path string) error {
	gopath := os.Getenv("GOPATH")
	repo := ponzuRepo
	local := filepath.Join(gopath, "src", filepath.Join(repo...))
	network := "https://" + strings.Join(repo, "/") + ".git"
	if !strings.HasPrefix(path, gopath) {
		path = filepath.Join(gopath, path)
	}

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
		fmt.Println("Couldn't clone from", local, "- trying network...")

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
	vendorPath := filepath.Join(path, "cmd", "ponzu", "vendor", "github.com", "ponzu-cms", "ponzu")
	err := os.MkdirAll(vendorPath, os.ModeDir|os.ModePerm)
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
	contentPath := filepath.Join(path, "content")
	err = os.Mkdir(contentPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func copyFileNoRoot(src, dst string) error {
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

		if info.IsDir() {
			// don't copy root directory
			if path == srcDir {
				return nil
			}

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

		err = copyFileNoRoot(path, dstDir)
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

func emptyDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
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
	dst := filepath.Join("cmd", "ponzu", "vendor", "github.com", "ponzu-cms", "ponzu", "content")
	err = emptyDir(dst)
	if err != nil {
		return err
	}
	err = copyFilesWarnConflicts(src, dst, []string{"doc.go"})
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
	buildOptions := []string{"build", "-o", "ponzu-server"}
	cmdBuildFiles := []string{"main.go", "options.go", "generate.go", "usage.go"}
	var cmdBuildFilePaths []string
	for _, file := range cmdBuildFiles {
		p := filepath.Join(pwd, "cmd", "ponzu", file)
		cmdBuildFilePaths = append(cmdBuildFilePaths, p)
	}

	build := exec.Command(gocmd, append(buildOptions, cmdBuildFilePaths...)...)
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

func copyAll(src, dst string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		sep := string(filepath.Separator)

		// base == the ponzu project dir + string(filepath.Separator)
		parts := strings.Split(src, sep)
		base := strings.Join(parts[:len(parts)-1], sep)
		base += sep

		target := filepath.Join(dst, path[len(base):])

		// if its a directory, make dir in dst
		if info.IsDir() {
			err := os.MkdirAll(target, os.ModeDir|os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			// if its a file, move file to dir of dst
			err = os.Rename(path, target)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
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
