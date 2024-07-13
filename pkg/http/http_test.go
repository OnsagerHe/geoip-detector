package http

import (
	"testing"

	"github.com/OnsagerHe/geoip-detector/pkg/utils"
)

func TestGetDomainFromURL(t *testing.T) {
	tests := []struct {
		name     string
		input    *utils.EndpointMetadata
		expected string
		hasError bool
	}{
		{
			name: "Valid URL with http",
			input: &utils.EndpointMetadata{
				Endpoint: "http://example.com/path",
			},
			expected: "example.com",
			hasError: false,
		},
		{
			name: "Valid URL with https",
			input: &utils.EndpointMetadata{
				Endpoint: "https://example.com/path",
			},
			expected: "example.com",
			hasError: false,
		},
		{
			name: "Valid URL with subdomain",
			input: &utils.EndpointMetadata{
				Endpoint: "https://sub.example.com/path",
			},
			expected: "sub.example.com",
			hasError: false,
		},
		{
			name: "Invalid URL",
			input: &utils.EndpointMetadata{
				Endpoint: "://invalid-url",
			},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := getDomainFromURL(tt.input)
			if (err != nil) != tt.hasError {
				t.Errorf("getDomainFromURL() error = %v, expected error = %v", err, tt.hasError)
				return
			}
			if tt.input.Host != tt.expected {
				t.Errorf("getDomainFromURL() got = %v, expected = %v", tt.input.Host, tt.expected)
			}
		})
	}
}

func TestParseHTTP(t *testing.T) {
	tests := []struct {
		name      string
		input     *utils.EndpointMetadata
		expected  *utils.EndpointMetadata
		expectErr bool
	}{
		{
			name: "HTTPS Endpoint",
			input: &utils.EndpointMetadata{
				Endpoint: "https://example.com",
			},
			expected: &utils.EndpointMetadata{
				Endpoint: "https://example.com",
				Prefix:   "https://",
				Port:     ":443",
			},
			expectErr: false,
		},
		{
			name: "HTTP Endpoint",
			input: &utils.EndpointMetadata{
				Endpoint: "http://example.com",
			},
			expected: &utils.EndpointMetadata{
				Endpoint: "http://example.com",
				Prefix:   "http://",
				Port:     ":80",
			},
			expectErr: false,
		},
		{
			name: "Invalid Endpoint",
			input: &utils.EndpointMetadata{
				Endpoint: "ftp://example.com",
			},
			expected: &utils.EndpointMetadata{
				Endpoint: "ftp://example.com",
			},
			expectErr: true,
		},
		{
			name: "Empty Endpoint",
			input: &utils.EndpointMetadata{
				Endpoint: "",
			},
			expected: &utils.EndpointMetadata{
				Endpoint: "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseHTTP(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("parseHTTP() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr && (tt.input.Prefix != tt.expected.Prefix || tt.input.Port != tt.expected.Port) {
				t.Errorf("parseHTTP() got = %v, expected = %v", tt.input, tt.expected)
			}
		})
	}
}
