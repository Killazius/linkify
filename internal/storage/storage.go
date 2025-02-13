package storage

import "errors"

var (
	ErrAliasExists   = errors.New("alias already exists")
	ErrURLNotFound   = errors.New("URL not found")
	ErrURLExists     = errors.New("URL already exists")
	ErrAliasNotFound = errors.New("alias not found")
)
