package utils

import (
	"os"
	"runtime"
	"strings"
)

type FS struct {
	isWin    bool
	Slash    string
	BasePath string
	baseLen  int
}

// @basePath If nil then use current working directory
func NewFS(basePath *[]string) FS {
	isWin := false
	slash := "/"
	if runtime.GOOS == "windows" {
		isWin = true
		slash = "\\"
	}
	bp, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if basePath != nil {
		bp = strings.Join(*basePath, slash)
	}
	fs := FS{
		isWin:    isWin,
		Slash:    slash,
		BasePath: bp,
		baseLen:  len(strings.Split(bp, slash)),
	}
	fs.mkdirIfNotExists(fs.Path(""), false)
	return fs
}

func (fs *FS) Path(path ...string) string {
	return fs.BasePath + fs.Slash + strings.Join(path, fs.Slash)
}

func (fs *FS) ToPath(path string) []string {
	parts := strings.Split(path, "/")
	relativePath := []string{}
	for i := range parts {
		part := parts[i]
		if part == "" {
			continue
		}
		relativePath = append(relativePath, part)
	}
	return relativePath
}

func (fs *FS) FileSize(path ...string) Result[int64] {
	fileInfo, err := os.Stat(fs.Path(path...))
	if err != nil {
		return Err[int64](err)
	}
	return Ok(fileInfo.Size())
}

func (fs *FS) Exists(path ...string) bool {
	_, err := os.Stat(fs.Path(path...))
	if err != nil {
		return false
	}
	return true
}

func (fs *FS) mkdirIfNotExists(path string, skipRoot bool) {
	parts := strings.Split(path, fs.Slash)
	startAt := 0
	if skipRoot {
		startAt = fs.baseLen
	}
	for i := startAt; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" {
			continue
		}
		path = strings.Join(parts[:i+1], fs.Slash)
		if !dirExists(path) {
			os.Mkdir(path, 0755)
		}
	}
}

func (fs *FS) Write(data []byte, path ...string) Result[bool] {
	aPath := fs.Path(path...)
	fs.mkdirIfNotExists(aPath, true)
	err := os.WriteFile(aPath, data, 0644)
	if err != nil {
		return Err[bool](err)
	}
	return Ok(true)
}

func (fs *FS) Append(data []byte, path ...string) Result[bool] {
	aPath := fs.Path(path...)
	fs.mkdirIfNotExists(aPath, true)
	f, err := os.OpenFile(aPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return Err[bool](err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return Err[bool](err)
	}
	return Ok(true)
}

func (fs *FS) Read(path ...string) Result[[]byte] {
	aPath := fs.Path(path...)
	data, err := os.ReadFile(aPath)
	if err != nil {
		return Err[[]byte](err)
	}
	return Ok(data)
}

func (fs *FS) ReadString(path ...string) Result[string] {
	result := fs.Read(path...)
	if result.Error != nil {
		return Err[string](result.Error)
	}
	return Ok(string(result.Value))
}

func (fs *FS) OpenFile(path ...string) Result[*os.File] {
	aPath := fs.Path(path...)
	file, err := os.Open(aPath)
	if err != nil {
		return Err[*os.File](err)
	}
	return Ok(file)
}

func (fs *FS) Delete(path ...string) Result[bool] {
	aPath := fs.Path(path...)
	err := os.Remove(aPath)
	if err != nil {
		return Err[bool](err)
	}
	return Ok(true)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return info.IsDir()
}

func (fs *FS) ListFiles(path ...string) Result[[]string] {
	aPath := fs.Path(path...)
	entries, err := os.ReadDir(aPath)
	if err != nil {
		return Err[[]string](err)
	}
	var files []string = []string{}
	for _, entry := range entries {
		nPath := path
		nPath = append(nPath, entry.Name())
		if entry.IsDir() {
			childFiles := fs.ListFiles(nPath...)
			if childFiles.Error != nil {
				return Err[[]string](childFiles.Error)
			}
			files = append(files, childFiles.Value...)
		} else {
			pathJoined := strings.Join(path, "/")
			if pathJoined == "" {
				files = append(files, entry.Name())
			} else {
				pathJoined := strings.TrimPrefix(pathJoined, "/")
				files = append(files, pathJoined+"/"+entry.Name())
			}
		}
	}
	return Ok(files)
}
