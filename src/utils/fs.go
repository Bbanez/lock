package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FS struct {
	BasePath string
}

// @basePath If nil then use current working directory
func NewFS(basePath *[]string) FS {
	bp, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	bp = filepath.Clean(bp)
	if basePath != nil && len(*basePath) > 0 {
		rel := filepath.Clean(filepath.Join((*basePath)...))
		if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			panic("input path escapes current working directory")
		}
		candidate := filepath.Join(bp, rel)
		relCheck, err := filepath.Rel(bp, candidate)
		if err != nil {
			panic(err)
		}
		if relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(os.PathSeparator)) {
			panic("input path escapes current working directory")
		}
		bp = candidate
	}
	fs := FS{
		BasePath: bp,
	}
	fmt.Println("PWD:", fs.Path())
	fs.mkdirIfNotExists(fs.Path())
	return fs
}

func (fs *FS) Path(path ...string) string {
	parts := append([]string{fs.BasePath}, path...)
	resolved := filepath.Clean(filepath.Join(parts...))
	rel, err := filepath.Rel(fs.BasePath, resolved)
	if err != nil {
		panic(err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		panic("resolved path escapes base directory")
	}
	return resolved
}

func (fs *FS) ToPath(path string) []string {
	parts := strings.Split(filepath.ToSlash(path), "/")
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

func (fs *FS) mkdirIfNotExists(path string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}
}

func (fs *FS) Write(data []byte, path ...string) Result[bool] {
	return fs.WriteWithMode(data, 0644, path...)
}

func (fs *FS) WriteWithMode(data []byte, mode os.FileMode, path ...string) Result[bool] {
	aPath := fs.Path(path...)
	fs.mkdirIfNotExists(aPath)
	err := os.WriteFile(aPath, data, mode.Perm())
	if err != nil {
		return Err[bool](err)
	}
	return Ok(true)
}

func (fs *FS) Append(data []byte, path ...string) Result[bool] {
	aPath := fs.Path(path...)
	fs.mkdirIfNotExists(aPath)
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

func (fs *FS) FileMode(path ...string) Result[os.FileMode] {
	fileInfo, err := os.Stat(fs.Path(path...))
	if err != nil {
		return Err[os.FileMode](err)
	}
	return Ok(fileInfo.Mode())
}

func (fs *FS) Delete(path ...string) Result[bool] {
	aPath := fs.Path(path...)
	err := os.Remove(aPath)
	if err != nil {
		return Err[bool](err)
	}
	return Ok(true)
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
			file := filepath.ToSlash(filepath.Join(append(path, entry.Name())...))
			files = append(files, file)
		}
	}
	return Ok(files)
}
