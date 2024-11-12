package ioutils

import (
	"encoding/csv"
	"os"
	"tp1/pkg/logs"
)

const filepath = "recovery.csv"

// File is a structure that encapsulates a CSV reader and the associated file.
type File struct {
	r    *csv.Reader
	file *os.File
}

func NewFile() (*File, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &File{
		r:    csv.NewReader(file),
		file: file,
	}, nil
}

func (r *File) ReadNext() ([]string, error) {
	return r.r.Read()
}

// Close closes the underlying file associated with the File and logs an error if the file cannot be closed.
func (r *File) Close() {
	if err := r.file.Close(); err != nil {
		logs.Logger.Errorf("error closing file: %v", err)
	}
}
