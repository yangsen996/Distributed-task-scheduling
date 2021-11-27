package utils

import (
	"io/ioutil"
	"os"
	"strings"
)

// Exist判断文件是否存在
func Exist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func toString(f string) (string, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func GetConfigInfo(f string) (string, error) {
	str, err := toString(f)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(str), nil
}
