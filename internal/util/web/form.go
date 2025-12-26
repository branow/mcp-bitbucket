package web

import (
	"io"
	"mime/multipart"
)

// MultipartForm represents a multipart/form-data form with multiple parts.
//
// Use MultipartForm when you need to send both text fields and file uploads
// in a single HTTP request.
//
// Example:
//
//	form := &MultipartForm{
//	  Parts: []FormPart{
//	    &TextField{Name: "title", Value: "My Document"},
//	    &FileField{Name: "file", Filename: "doc.pdf", Reader: fileReader},
//	  },
//	}
type MultipartForm struct {
	// Parts is a list of form parts (text fields or file fields).
	Parts []FormPart
}

// FormPart represents a single part in a multipart form.
// It can be either a TextField or a FileField.
type FormPart interface {
	// Write writes this form part to the multipart writer.
	Write(w *multipart.Writer) error
}

// TextField represents a text field in a multipart form.
type TextField struct {
	// Name is the form field name.
	Name string
	// Value is the text value of the field.
	Value string
}

// Write writes this text field to the multipart writer.
func (f *TextField) Write(w *multipart.Writer) error {
	return w.WriteField(f.Name, f.Value)
}

// FileField represents a file upload field in a multipart form.
type FileField struct {
	// Name is the form field name.
	Name string
	// Filename is the name of the file being uploaded.
	Filename string
	// Reader provides the file content.
	Reader io.Reader
}

// Write writes this file field to the multipart writer.
func (f *FileField) Write(w *multipart.Writer) error {
	part, err := w.CreateFormFile(f.Name, f.Filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f.Reader)
	return err
}
