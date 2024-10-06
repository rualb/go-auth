package toolconfig

import (
	"encoding/json"
	"fmt"
	"go-auth/internal/tool/toolhttp"
	xlog "go-auth/internal/tool/toollog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func LoadConfig(cnfPtr any, dir string, fileName string) error {

	xlog.Info("Loading config from: %v", dir)

	isHTTP := strings.HasPrefix(dir, "http")

	if isHTTP {

		err := fromURL(cnfPtr, dir, fileName)
		if err != nil {
			return err
		}

	} else {
		err := fromFile(cnfPtr, dir, fileName)
		if err != nil {
			return err
		}
	}

	return nil
}

// fromFile errIfNotExists argument soft binding, no error if file not exists
func fromFile(cnfPtr any, dir string, file string) error {

	if file == "" {
		return nil
	}

	if !strings.HasSuffix(file, ".json") {
		return fmt.Errorf("error file not match  *.json: %v", file)
	}

	fullPath, err := filepath.Abs(filepath.Join(dir, file))

	if err != nil {
		return err
	}

	fullPath = filepath.Clean(fullPath)

	data, err := os.ReadFile(fullPath)

	if err != nil {
		return fmt.Errorf("error with file %v: %v", fullPath, err)
	}

	xlog.Info("Loading config from file: %v", fullPath)

	err = fromJSON(cnfPtr, string(data))

	if err != nil {
		return err
	}

	return nil
}

// fromURL errIfNotExists argument soft binding, no error if file not exists
func fromURL(cnfPtr any, dir string, file string) error {

	if file == "" {
		return nil
	}

	if !strings.HasSuffix(file, ".json") {
		return fmt.Errorf("error file not match  *.json: %v", file)
	}

	fullPath := dir + "/" + file

	_, err := url.Parse(fullPath)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	// fmt.Println("Reading config from file: ", file)

	data, err := toolhttp.GetBytes(fullPath, nil, nil)

	if err != nil {
		return fmt.Errorf("error with file %v: %v", fullPath, err)
	}

	xlog.Info("Loading config from file: %v", fullPath)

	err = fromJSON(cnfPtr, string(data))
	if err != nil {
		return err
	}

	return nil
}

func fromJSON(cnfPtr any, data string) error {

	if data == "" {
		return nil
	}

	err := json.Unmarshal([]byte(data), cnfPtr)

	if err != nil {
		return err
	}

	return nil
}
