package mcp

import "errors"

var (
	// ErrInvalidPort is returned when the port number is invalid
	ErrInvalidPort = errors.New("invalid port number")
	
	// ErrMissingTLSCert is returned when TLS is enabled but no certificate path is provided
	ErrMissingTLSCert = errors.New("TLS certificate path is required when TLS is enabled")
	
	// ErrMissingTLSKey is returned when TLS is enabled but no private key path is provided
	ErrMissingTLSKey = errors.New("TLS private key path is required when TLS is enabled")
	
	// ErrResourceNotFound is returned when a requested resource is not found
	ErrResourceNotFound = errors.New("resource not found")
	
	// ErrToolNotFound is returned when a requested tool is not found
	ErrToolNotFound = errors.New("tool not found")
	
	// ErrInvalidToolArguments is returned when tool arguments are invalid
	ErrInvalidToolArguments = errors.New("invalid tool arguments")
)
