package utils

import (
	"os"
	"path/filepath"
	"strings"
)

type Args struct {
	Input        []string
	Lock         bool
	Pass         string
	ProjectBuild bool
	Release      bool
}

func GetArgs() Args {
	args := Args{
		Input: []string{},
		Pass:  "",
		Lock:  true,
	}
	rawArgs := os.Args[1:]
	i := 0
	for i < len(rawArgs) {
		if i+1 >= len(rawArgs) {
			break
		}
		value := rawArgs[i]
		switch value {
		case "-i":
			clean := filepath.Clean(rawArgs[i+1])
			if clean == "." {
				args.Input = []string{}
				break
			}
			if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
				panic("-i path cannot escape current working directory")
			}
			args.Input = strings.Split(filepath.ToSlash(clean), "/")
		case "-p":
			args.Pass = rawArgs[i+1]
		case "-u":
			args.Lock = false
		case "-r":
			args.Release = true
		case "-project-build":
			args.ProjectBuild = true
		}
		i += 2
	}
	return args
}
