package web_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextField_Write(t *testing.T) {
	tests := []struct {
		name              string
		field             web.TextField
		expectedInContent []string
	}{
		{
			name: "normal field with value",
			field: web.TextField{
				Name:  "username",
				Value: "john_doe",
			},
			expectedInContent: []string{"username", "john_doe"},
		},
		{
			name: "empty value",
			field: web.TextField{
				Name:  "field",
				Value: "",
			},
			expectedInContent: []string{"field"},
		},
		{
			name: "special characters",
			field: web.TextField{
				Name:  "message",
				Value: "Hello\nWorld\r\n!@#$%^&*()",
			},
			expectedInContent: []string{"message", "Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			writer := multipart.NewWriter(buf)

			err := tt.field.Write(writer)
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expectedInContent {
				assert.Contains(t, content, expected)
			}
		})
	}
}

func TestFileField_Write(t *testing.T) {
	tests := []struct {
		name              string
		field             web.FileField
		expectedInContent []string
		minContentLength  int
	}{
		{
			name: "normal file with content",
			field: web.FileField{
				Name:     "document",
				Filename: "test.txt",
				Reader:   strings.NewReader("This is the file content"),
			},
			expectedInContent: []string{"document", "test.txt", "This is the file content", "Content-Disposition", `filename="test.txt"`},
		},
		{
			name: "empty file",
			field: web.FileField{
				Name:     "empty",
				Filename: "empty.txt",
				Reader:   strings.NewReader(""),
			},
			expectedInContent: []string{"empty", "empty.txt"},
		},
		{
			name: "binary content",
			field: web.FileField{
				Name:     "binary",
				Filename: "data.bin",
				Reader:   bytes.NewReader([]byte{0x00, 0xFF, 0xAB, 0xCD, 0xEF}),
			},
			expectedInContent: []string{"binary", "data.bin"},
		},
		{
			name: "large file (1MB)",
			field: web.FileField{
				Name:     "large",
				Filename: "large.txt",
				Reader:   strings.NewReader(strings.Repeat("A", 1024*1024)),
			},
			expectedInContent: []string{"large", "large.txt"},
			minContentLength:  1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			writer := multipart.NewWriter(buf)

			err := tt.field.Write(writer)
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			content := buf.String()
			for _, expected := range tt.expectedInContent {
				assert.Contains(t, content, expected)
			}

			if tt.minContentLength > 0 {
				assert.Greater(t, len(content), tt.minContentLength)
			}
		})
	}
}

func TestFileField_Write_ReadError(t *testing.T) {
	errorReader := &errorReader{err: io.ErrUnexpectedEOF}
	field := &web.FileField{
		Name:     "error",
		Filename: "error.txt",
		Reader:   errorReader,
	}

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	err := field.Write(writer)
	require.Error(t, err)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestMultipartForm_MixedFields(t *testing.T) {
	form := &web.MultipartForm{
		Parts: []web.FormPart{
			&web.TextField{Name: "title", Value: "My Document"},
			&web.TextField{Name: "description", Value: "A test document"},
			&web.FileField{Name: "file", Filename: "document.pdf", Reader: strings.NewReader("PDF content")},
			&web.TextField{Name: "author", Value: "John Doe"},
		},
	}

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	for _, part := range form.Parts {
		err := part.Write(writer)
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	content := buf.String()
	assert.Contains(t, content, "title")
	assert.Contains(t, content, "My Document")
	assert.Contains(t, content, "description")
	assert.Contains(t, content, "A test document")
	assert.Contains(t, content, "file")
	assert.Contains(t, content, "document.pdf")
	assert.Contains(t, content, "PDF content")
	assert.Contains(t, content, "author")
	assert.Contains(t, content, "John Doe")
}

func TestMultipartForm_ParseRoundTrip(t *testing.T) {
	form := &web.MultipartForm{
		Parts: []web.FormPart{
			&web.TextField{Name: "name", Value: "Test User"},
			&web.FileField{Name: "avatar", Filename: "avatar.png", Reader: strings.NewReader("image data")},
		},
	}

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()

	for _, part := range form.Parts {
		err := part.Write(writer)
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	reader := multipart.NewReader(buf, boundary)

	part1, err := reader.NextPart()
	require.NoError(t, err)
	assert.Equal(t, "name", part1.FormName())
	value1, err := io.ReadAll(part1)
	require.NoError(t, err)
	assert.Equal(t, "Test User", string(value1))
	part1.Close()

	part2, err := reader.NextPart()
	require.NoError(t, err)
	assert.Equal(t, "avatar", part2.FormName())
	assert.Equal(t, "avatar.png", part2.FileName())
	value2, err := io.ReadAll(part2)
	require.NoError(t, err)
	assert.Equal(t, "image data", string(value2))
	part2.Close()

	_, err = reader.NextPart()
	assert.Equal(t, io.EOF, err)
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
