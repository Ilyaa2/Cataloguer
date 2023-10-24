package util

import (
	"errors"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func StripFileNameFromPath(path string) string {
	s := path[strings.LastIndex(path, "/")+1:]
	ans, _, _ := strings.Cut(s, ".")
	return ans
}

func NumberOfUsersDirectory(url string) (int, error) {
	i := strings.LastIndex(url, "user")
	if i == -1 {
		return 0, errors.New("no id found")
	}
	i += 4
	id := ""
	for i != len(url) && string(url[i]) != "/" {
		id += string(url[i])
		i++
	}
	ans, err := strconv.Atoi(id)
	if id == "" || err != nil {
		return 0, errors.New("no id found")
	}
	return ans, nil
}

func StripUrlFromFilePath(fullPath string, basePath string) string {
	_, ans, _ := strings.Cut(filepath.ToSlash(fullPath), filepath.ToSlash(basePath))
	return ans
}
