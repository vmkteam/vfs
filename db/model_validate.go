//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package db

import (
	"unicode/utf8"
)

const (
	ErrEmptyValue = "empty"
	ErrMaxLength  = "len"
	ErrWrongValue = "value"
)

func (vf VfsFile) Validate() (errors map[string]string, valid bool) {
	errors = map[string]string{}

	if vf.ID == 0 {
		errors[Columns.VfsFile.ID] = ErrEmptyValue
	}

	if utf8.RuneCountInString(vf.Title) > 255 {
		errors[Columns.VfsFile.Title] = ErrMaxLength
	}

	if utf8.RuneCountInString(vf.Path) > 255 {
		errors[Columns.VfsFile.Path] = ErrMaxLength
	}

	if utf8.RuneCountInString(vf.MimeType) > 255 {
		errors[Columns.VfsFile.MimeType] = ErrMaxLength
	}

	return errors, len(errors) == 0
}

func (vf VfsFolder) Validate() (errors map[string]string, valid bool) {
	errors = map[string]string{}

	if vf.ID == 0 {
		errors[Columns.VfsFolder.ID] = ErrEmptyValue
	}

	if utf8.RuneCountInString(vf.Title) > 255 {
		errors[Columns.VfsFolder.Title] = ErrMaxLength
	}

	return errors, len(errors) == 0
}
