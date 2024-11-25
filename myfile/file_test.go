package myfile

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddGetFileSqlite(t *testing.T) {
	f, err := os.CreateTemp("", "sample")
	assert.NoError(t, err, "failed to create temp file")
	defer os.Remove(f.Name())
	db, err := NewSQLiteController(f.Name())
	assert.NoError(t, err, "failed to create sqlite controller")
	fileJSON := ` {
    "name": "file.txt",
    "uploaded_by": "File Owner",
    "email": "test@test.com",
    "data": "aGVsbG8gd29ybGQK"
  } `
	mf := MyFile{}
	err = json.Unmarshal([]byte(fileJSON), &mf)
	assert.NoError(t, err, "failed to create file from json")
	err = db.AddFile(&mf)
	assert.NoError(t, err, "failed to add file")
	file, err := db.GetFileByID(1)
	assert.NoError(t, err, "failed to get file")
	assert.Equal(t, "hello world\n", string(file.Data))
}
