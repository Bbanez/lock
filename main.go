package main

import (
	"fmt"

	"os"
	"os/exec"
	"strings"

	"github.com/bbanez/lock/src/security"
	"github.com/bbanez/lock/src/utils"
	"golang.org/x/term"
)

const Version = "v0.1.2"

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = fmt.Errorf("operation failed: %w", v)
			default:
				err = fmt.Errorf("operation failed: %v", v)
			}
		}
	}()

	args := utils.GetArgs()
	fs := utils.NewFS(&args.Input)
	if args.ProjectBuild {
		return BuildProject(args)
	}
	filesRes := fs.ListFiles()
	if filesRes.Error != nil {
		return fmt.Errorf("failed to list input files: %w", filesRes.Error)
	}
	files := filesRes.Value
	if args.Lock {
		if args.Pass == "" {
			fmt.Print("Enter password: ")
			pass, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			args.Pass = string(pass)
		}
		for i := range files {
			file := files[i]
			fmt.Printf("Encrypting file %s...", file)
			filePath := strings.Split(file, "/")
			modeRes := fs.FileMode(filePath...)
			if modeRes.Error != nil {
				fmt.Printf("[ERROR] reading mode for file %s: %s\n", file, modeRes.Error)
				continue
			}
			fileRes := fs.Read(filePath...)
			if fileRes.Error != nil {
				fmt.Printf("[ERROR] reading file %s: %s\n", file, fileRes.Error)
				continue
			}
			content := fileRes.Value
			encrypted, err := security.Encrypt(args.Pass, content, true)
			if err != nil {
				fmt.Printf("[ERROR] encrypting file %s: %s\n", file, err)
				continue
			}
			writeRes := fs.WriteWithMode(encrypted, modeRes.Value, filePath...)
			if writeRes.Error != nil {
				fmt.Printf("[ERROR] writing file %s: %s\n", file, writeRes.Error)
				continue
			}
			fmt.Printf("Done\n")
		}
		return nil
	} else {
		if args.Pass == "" {
			fmt.Print("Enter password: ")
			pass, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			args.Pass = string(pass)
		}
		for i := range files {
			file := files[i]
			fmt.Printf("Decrypting file %s...", file)
			filePath := strings.Split(file, "/")
			modeRes := fs.FileMode(filePath...)
			if modeRes.Error != nil {
				fmt.Printf("[ERROR] reading mode for file %s: %s\n", file, modeRes.Error)
				continue
			}
			fileRes := fs.Read(filePath...)
			if fileRes.Error != nil {
				fmt.Printf("[ERROR] reading file %s: %s\n", file, fileRes.Error)
				continue
			}
			content := fileRes.Value
			decrypted, err := security.Decrypt(args.Pass, content, true)
			if err != nil {
				fmt.Printf("[ERROR] decrypting file %s: %s\n", file, err)
				continue
			}
			writeRes := fs.WriteWithMode(decrypted, modeRes.Value, filePath...)
			if writeRes.Error != nil {
				fmt.Printf("[ERROR] writing file %s: %s\n", file, writeRes.Error)
				continue
			}
			fmt.Printf("Done\n")
		}
		return nil
	}
}

type BuildOutputs struct {
	Name     string
	Platform string
	Arch     string
}

func BuildProject(args utils.Args) error {
	fmt.Println("Building project...")
	buildOutputs := []BuildOutputs{
		{
			Name:     "Linux",
			Platform: "linux",
			Arch:     "amd64",
		},
		{
			Name:     "Linux",
			Platform: "linux",
			Arch:     "arm64",
		},
		{
			Name:     "Windows",
			Platform: "windows",
			Arch:     "amd64",
		},
		{
			Name:     "MacOS",
			Platform: "darwin",
			Arch:     "arm64",
		},
	}
	for i := range buildOutputs {
		output := buildOutputs[i]
		fmt.Printf(
			"Build %s %s binary...",
			output.Name, output.Arch,
		)
		ext := ""
		if output.Platform == "windows" {
			ext = ".exe"
		}
		cmd := exec.Command(
			"go",
			"build",
			"-o",
			fmt.Sprintf(
				"lock_release_%s_%s_%s%s",
				strings.ReplaceAll(Version, ".", "-"),
				output.Platform, output.Arch,
				ext,
			),
		)
		cmd.Env = append(os.Environ(),
			"GOOS="+output.Platform,
			"GOARCH="+output.Arch,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("build failed for %s/%s: %w", output.Platform, output.Arch, err)
		}
		fmt.Printf(" Done\n")
	}
	if !args.Release {
		return nil
	}
	cmd := exec.Command("git", "show-ref", "--tags")
	tagsStr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to read git tags: %w", err)
	}
	tagFound := false
	tags := strings.Split(string(tagsStr), "\n")
	for i := range tags {
		tag := tags[i]
		if strings.Contains(tag, "tags/"+Version) {
			tagFound = true
			break
		}
	}
	if !tagFound {
		fmt.Println("Creating a git tag")
		cmd := exec.Command("git", "tag", Version)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to create git tag %s: %w", Version, err)
		}
	} else {
		fmt.Println("Git tag already exists")
	}
	cmd = exec.Command("git", "push", "origin", Version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to push git tag %s: %w", Version, err)
	}
	cmd = exec.Command(
		"gh", "release", "create", Version,
		"lock_release_*",
		"--title", "\"Release "+Version+"\"",
		"--generate-notes",
	)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create GitHub release %s: %w", Version, err)
	}
	fmt.Printf("Cleanup...")
	for i := range buildOutputs {
		output := buildOutputs[i]
		ext := ""
		if output.Platform == "windows" {
			ext = ".exe"
		}
		filename :=
			fmt.Sprintf(
				"lock_release_%s_%s_%s%s",
				strings.ReplaceAll(Version, ".", "-"),
				output.Platform, output.Arch,
				ext,
			)
		if _, err := os.Stat(filename); err == nil {
			err := os.Remove(filename)
			if err != nil {
				fmt.Println(err)
			}
		} else if os.IsNotExist(err) {
			continue
		} else {
			continue
		}
	}
	fmt.Printf(" Done\n")
	return nil
}
