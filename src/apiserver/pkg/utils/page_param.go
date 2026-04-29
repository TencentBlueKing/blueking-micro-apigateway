// Package utils provides shared helpers and small cross-package data types.
package utils

// PageParam carries offset/limit pagination parameters shared across packages.
type PageParam struct {
	Offset int
	Limit  int
}
