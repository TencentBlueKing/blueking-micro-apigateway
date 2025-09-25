package filex

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
)

// 辅助函数：模拟生成 multipart.FileHeader
func createTestFileHeader(content string) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.json")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write([]byte(content)); err != nil {
		return nil, err
	}
	writer.Close()
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	reader, err := req.MultipartReader()
	if err != nil {
		return nil, err
	}
	form, err := reader.ReadForm(1024)
	if err != nil {
		return nil, err
	}
	files := form.File["file"]
	if len(files) == 0 {
		return nil, errors.New("file not found")
	}
	return files[0], nil
}

// 测试正常流程
func TestReadFileToObject_Success(t *testing.T) {
	content := `{"name": "test", "value": 123}`
	fileHeader, err := createTestFileHeader(content)
	if err != nil {
		t.Fatalf("创建 FileHeader 失败: %v", err)
	}
	var obj struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	err = ReadFileToObject(fileHeader, &obj)
	if err != nil {
		t.Fatalf("预期无错误，实际出错: %v", err)
	}
	if obj.Name != "test" || obj.Value != 123 {
		t.Errorf("对象反序列化异常，预期 Name=test, Value=123，实际 Name=%s, Value=%d", obj.Name, obj.Value)
	}
}

// 测试文件打开失败
func TestReadFileToObject_OpenFileError(t *testing.T) {
	// 构造一个无效的 FileHeader（模拟无法打开文件的情况）
	invalidFileHeader := &multipart.FileHeader{
		Filename: "invalid.txt",
		Header:   nil, // Header 为空将导致 Open 方法失败
	}
	var obj interface{}
	err := ReadFileToObject(invalidFileHeader, &obj)
	if err == nil || !strings.Contains(err.Error(), "open file failed") {
		t.Errorf("预期错误包含 'open file failed'，实际错误: %v", err)
	}
}

// 测试 JSON 反序列化失败
func TestReadFileToObject_UnmarshalError(t *testing.T) {
	content := `{"name": "test", "value": "invalid_int"}` // value 应为数字
	fileHeader, err := createTestFileHeader(content)
	if err != nil {
		t.Fatalf("创建 FileHeader 失败: %v", err)
	}
	var obj struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	err = ReadFileToObject(fileHeader, &obj)
	if err == nil || !strings.Contains(err.Error(), "unmarshal file failed") {
		t.Errorf("预期错误包含 'unmarshal file failed'，实际错误: %v", err)
	}
}
