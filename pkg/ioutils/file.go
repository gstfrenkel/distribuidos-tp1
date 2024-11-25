package ioutils

import (
	"encoding/csv"
	"os"
	"tp1/pkg/logs"
)

const (
	filePath = "recovery.csv"
	fileMode = 0666
)

// File is a structure that encapsulates a CSV reader and the associated file.
type File struct {
	r    *csv.Reader
	w    *csv.Writer
	file *os.File
}

func NewFile() (*File, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, fileMode)
	if err != nil {
		return nil, err
	}
	return &File{
		r:    csv.NewReader(file),
		w:    csv.NewWriter(file),
		file: file,
	}, nil
}

func (r *File) Read() ([]string, error) {
	return r.r.Read()
}

func (r *File) Write(record []string) error {
	defer r.w.Flush()
	return r.w.Write(record)
}

// Close closes the underlying file associated with the File and logs an error if the file cannot be closed.
func (r *File) Close() {
	if err := r.file.Close(); err != nil {
		logs.Logger.Errorf("error closing file: %v", err)
	}
}
