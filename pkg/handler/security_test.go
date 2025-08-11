package handler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeMessageContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		contentType string
		expected    string
		expectError bool
	}{
		{
			name:        "Valid plain text",
			content:     "Hello world",
			contentType: "text/plain",
			expected:    "Hello world",
			expectError: false,
		},
		{
			name:        "Valid markdown text",
			content:     "# Hello World",
			contentType: "text/markdown",
			expected:    "# Hello World",
			expectError: false,
		},
		{
			name:        "HTML in markdown gets escaped",
			content:     "<script>alert('xss')</script>",
			contentType: "text/markdown",
			expected:    "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
			expectError: false,
		},
		{
			name:        "HTML in plain text stays as-is",
			content:     "<b>Bold text</b>",
			contentType: "text/plain",
			expected:    "<b>Bold text</b>",
			expectError: false,
		},
		{
			name:        "Unicode content is valid",
			content:     "Hello 世界 🌍",
			contentType: "text/plain",
			expected:    "Hello 世界 🌍",
			expectError: false,
		},
		{
			name:        "Empty content",
			content:     "",
			contentType: "text/plain",
			expected:    "",
			expectError: false,
		},
		{
			name:        "Content exceeding max length",
			content:     strings.Repeat("A", maxMessageLength+1),
			contentType: "text/plain",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Content at max length boundary",
			content:     strings.Repeat("A", maxMessageLength),
			contentType: "text/plain",
			expected:    strings.Repeat("A", maxMessageLength),
			expectError: false,
		},
		{
			name:        "Invalid UTF-8 sequence",
			content:     string([]byte{0xff, 0xfe, 0xfd}),
			contentType: "text/plain",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizeMessageContent(tt.content, tt.contentType)
			
			if tt.expectError {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
			} else {
				require.NoError(t, err, "Unexpected error for test case: %s", tt.name)
				assert.Equal(t, tt.expected, result, "Unexpected result for test case: %s", tt.name)
			}
		})
	}
}

func TestValidateChannelIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		channelID   string
		expectError bool
	}{
		{
			name:        "Valid channel ID with C prefix",
			channelID:   "C1234567890",
			expectError: false,
		},
		{
			name:        "Valid DM ID with D prefix",
			channelID:   "D1234567890",
			expectError: false,
		},
		{
			name:        "Valid group ID with G prefix",
			channelID:   "G1234567890",
			expectError: false,
		},
		{
			name:        "Valid channel name with # prefix",
			channelID:   "#general",
			expectError: false,
		},
		{
			name:        "Valid user mention with @ prefix",
			channelID:   "@username",
			expectError: false,
		},
		{
			name:        "Invalid prefix",
			channelID:   "X1234567890",
			expectError: true,
		},
		{
			name:        "Too short ID",
			channelID:   "C123",
			expectError: true,
		},
		{
			name:        "Lowercase letters in ID",
			channelID:   "C123456789a",
			expectError: true,
		},
		{
			name:        "Empty string",
			channelID:   "",
			expectError: true, // Empty should be invalid
		},
		{
			name:        "Too long channel identifier",
			channelID:   strings.Repeat("A", maxChannelNameLength+1),
			expectError: true,
		},
		{
			name:        "Invalid UTF-8 in channel identifier",
			channelID:   string([]byte{0xff, 0xfe, 0xfd}),
			expectError: true,
		},
		{
			name:        "Channel ID with special characters",
			channelID:   "C123456789@",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChannelIdentifier(tt.channelID)
			
			if tt.expectError {
				assert.Error(t, err, "Expected error for channel ID: %s", tt.channelID)
			} else {
				assert.NoError(t, err, "Unexpected error for channel ID: %s", tt.channelID)
			}
		})
	}
}

func TestValidateThreadTimestamp(t *testing.T) {
	tests := []struct {
		name        string
		threadTs    string
		expectError bool
	}{
		{
			name:        "Valid timestamp",
			threadTs:    "1234567890.123456",
			expectError: false,
		},
		{
			name:        "Empty timestamp (allowed)",
			threadTs:    "",
			expectError: false,
		},
		{
			name:        "Too few microseconds",
			threadTs:    "1234567890.12345",
			expectError: true,
		},
		{
			name:        "Too many microseconds",
			threadTs:    "1234567890.1234567",
			expectError: true,
		},
		{
			name:        "Too few seconds digits",
			threadTs:    "123456789.123456",
			expectError: true,
		},
		{
			name:        "Too many seconds digits",
			threadTs:    "12345678901.123456",
			expectError: true,
		},
		{
			name:        "Missing dot",
			threadTs:    "1234567890123456",
			expectError: true,
		},
		{
			name:        "Non-numeric seconds",
			threadTs:    "abcdefghij.123456",
			expectError: true,
		},
		{
			name:        "Non-numeric microseconds",
			threadTs:    "1234567890.abcdef",
			expectError: true,
		},
		{
			name:        "Multiple dots",
			threadTs:    "1234567890.123.456",
			expectError: true,
		},
		{
			name:        "Too long timestamp",
			threadTs:    strings.Repeat("1", maxThreadTsLength+1),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateThreadTimestamp(tt.threadTs)
			
			if tt.expectError {
				assert.Error(t, err, "Expected error for timestamp: %s", tt.threadTs)
			} else {
				assert.NoError(t, err, "Unexpected error for timestamp: %s", tt.threadTs)
			}
		})
	}
}

func TestIsChannelAllowed(t *testing.T) {
	tests := []struct {
		name      string
		channel   string
		envVar    string
		expected  bool
	}{
		{
			name:     "No config set - deny by default",
			channel:  "C1234567890",
			envVar:   "",
			expected: false,
		},
		{
			name:     "Config set to true - allow all",
			channel:  "C1234567890",
			envVar:   "true",
			expected: true,
		},
		{
			name:     "Config set to 1 - allow all",
			channel:  "C1234567890",
			envVar:   "1",
			expected: true,
		},
		{
			name:     "Channel in whitelist",
			channel:  "C1234567890",
			envVar:   "C1234567890,D0987654321",
			expected: true,
		},
		{
			name:     "Channel not in whitelist",
			channel:  "C9999999999",
			envVar:   "C1234567890,D0987654321",
			expected: false,
		},
		{
			name:     "Channel in blacklist",
			channel:  "C1234567890",
			envVar:   "!C1234567890,!D0987654321",
			expected: false,
		},
		{
			name:     "Channel not in blacklist",
			channel:  "C9999999999",
			envVar:   "!C1234567890,!D0987654321",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for the test
			if tt.envVar != "" {
				t.Setenv("SLACK_MCP_ADD_MESSAGE_TOOL", tt.envVar)
			}
			
			result := isChannelAllowed(tt.channel)
			assert.Equal(t, tt.expected, result, "Unexpected result for channel: %s with config: %s", tt.channel, tt.envVar)
		})
	}
}

// Benchmark tests for performance-critical security functions
func BenchmarkSanitizeMessageContent(b *testing.B) {
	content := strings.Repeat("Hello <script>alert('test')</script> world! ", 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sanitizeMessageContent(content, "text/markdown")
	}
}

func BenchmarkValidateChannelIdentifier(b *testing.B) {
	channelID := "C1234567890"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateChannelIdentifier(channelID)
	}
}

func BenchmarkValidateThreadTimestamp(b *testing.B) {
	threadTs := "1234567890.123456"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateThreadTimestamp(threadTs)
	}
}