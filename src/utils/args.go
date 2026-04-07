package utils

import (
	"os"
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
			args.Input = strings.Split(rawArgs[i+1], "/")
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
