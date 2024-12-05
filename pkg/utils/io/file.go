package io

import (
	"encoding/csv"
	"os"

	"tp1/pkg/logs"
)

const (
	fileMode = 0666
)

// File is a structure that encapsulates a CSV reader and the associated file.
type File struct {
	r    *csv.Reader
	w    *csv.Writer
	file *os.File
}

// NewFile creates a CSV file reader and writer for the specified filePath.
func NewFile(filePath string) (*File, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, fileMode)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1

	return &File{
		r:    r,
		w:    csv.NewWriter(file),
		file: file,
	}, nil
}

// Read reads the next line.
func (r *File) Read() ([]string, error) {
	return r.r.Read()
}

// Write writes a record into the file. Upon finishing, the writer's buffer gets flushed.
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

// Overwrite overwrites the file with the given strings.
func (r *File) Overwrite(strings []string) error {
	if err := r.file.Truncate(0); err != nil {
		return err
	}
	if _, err := r.file.Seek(0, 0); err != nil {
		return err
	}

	return r.Write(strings)
}
