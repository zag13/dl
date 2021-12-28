package core

import "io/ioutil"

var existedFiles = make([]string, 0)

func IsExist(key string, path string) (ok bool, err error) {
	if len(existedFiles) == 0 {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return false, err
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			existedFiles = append(existedFiles, file.Name())
		}
	}

	for _, file := range existedFiles {
		if key == file {
			return true, nil
		}
	}

	return false, nil
}
