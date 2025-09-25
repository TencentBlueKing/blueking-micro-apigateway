package filex

import (
	"bytes"
	"encoding/json"
	"mime/multipart"

	"github.com/pkg/errors"
)

func ReadFileToObject(fileHeader *multipart.FileHeader, obj interface{}) error {
	file, err := fileHeader.Open()
	if err != nil {
		return errors.Wrap(err, "open file failed")
	}
	defer file.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return errors.Wrap(err, "read file failed")
	}
	rawData := buf.Bytes()
	if err := json.Unmarshal(rawData, obj); err != nil {
		return errors.Wrap(err, "unmarshal file failed")
	}
	return nil
}
