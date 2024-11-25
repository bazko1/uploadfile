package myfile

import (
	"fmt"

	"gorm.io/driver/sqlite" // Sqlite driver based on CGO

	// "github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
)

// TODO :
type MyFile struct {
	gorm.Model
	FileMeta `gorm:"embedded"`
	Data     []byte `json:"data,omitempty"`
}

type FileMeta struct {
	Name       string `json:"name,omitempty"`
	UploadedBy string `json:"uploaded_by,omitempty"`
	Email      string `json:"email,omitempty"`
}

// SQLiteController is a controller with sqlite backend.
type FileController struct {
	DataSource string
	db         *gorm.DB
}

func NewSQLiteController(dataSource string) (FileController, error) {
	db, err := gorm.Open(sqlite.Open(dataSource), &gorm.Config{})
	if err != nil {
		return FileController{}, fmt.Errorf("failed to open db connection: %w", err)
	}

	err = db.AutoMigrate(&MyFile{})
	if err != nil {
		return FileController{}, fmt.Errorf("failed to automigrate db: %w", err)
	}

	return FileController{
		DataSource: dataSource,
		db:         db,
	}, nil
}

// adds single file to db
func (fc FileController) AddFile(file *MyFile) error {
	if result := fc.db.Create(file); result.Error != nil {
		return fmt.Errorf("failed to add file error: %w", result.Error)
	}
	return nil
}

// lists files without filling the data field
func (fc FileController) GetFilesMeta() (fmetas []MyFile, err error) {
	if result := fc.db.Select("id", "name", "uploaded_by", "email").Find(&fmetas); result.Error != nil {
		return nil, fmt.Errorf("failed to list files error: %w", result.Error)
	}

	return fmetas, nil
}

// get file by id
func (fc FileController) GetFileByID(id int) (file MyFile, err error) {
	if result := fc.db.First(&file, id); result.Error != nil {
		return file, fmt.Errorf("failed to list files error: %w", result.Error)
	}

	return file, nil
}
