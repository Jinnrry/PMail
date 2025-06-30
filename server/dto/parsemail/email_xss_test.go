package parsemail

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/emersion/go-message"
)

// XSS Test Cases
func TestSanitizeHTML(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple script tag",
			input:    `<script>alert('xss')</script>`,
			expected: ``,
		},
		{
			name:     "Image with onerror - should remove dangerous attributes",
			input:    `<img src=x onerror=alert('xss')>`,
			expected: ``, // bluemonday UGCPolicy removes img tags with dangerous attributes
		},
		{
			name:     "SVG with onload",
			input:    `<svg onload=alert('xss')></svg>`,
			expected: ``,
		},
		{
			name:     "Iframe with javascript",
			input:    `<iframe src=javascript:alert('xss')></iframe>`,
			expected: ``,
		},
		{
			name:     "Safe HTML content",
			input:    `<p>Hello <strong>World</strong>!</p>`,
			expected: `<p>Hello <strong>World</strong>!</p>`,
		},
		{
			name:     "Link with javascript protocol - removes dangerous href",
			input:    `<a href="javascript:alert('xss')">Click me</a>`,
			expected: `Click me`, // bluemonday removes entire link when href is dangerous
		},
		{
			name:     "Style tag with javascript",
			input:    `<style>body{background:url(javascript:alert('xss'))}</style>`,
			expected: ``,
		},
		{
			name:     "Object tag",
			input:    `<object data="javascript:alert('xss')"></object>`,
			expected: ``,
		},
		{
			name:     "Form tag",
			input:    `<form><input type="text" name="test"></form>`,
			expected: ``,
		},
		{
			name:     "Mixed content with XSS - keeps safe div tags",
			input:    `<p>Safe content</p><script>alert('xss')</script><div>More safe content</div>`,
			expected: `<p>Safe content</p><div>More safe content</div>`, // div is now allowed, script removed
		},
		{
			name:     "Event handlers in various tags - removes handlers but keeps tag",
			input:    `<div onclick="alert('xss')">Click me</div>`,
			expected: `<div>Click me</div>`, // div is kept, onclick handler removed
		},
		{
			name:     "Data URLs with javascript - removes dangerous link",
			input:    `<a href="data:text/html,<script>alert('xss')</script>">Link</a>`,
			expected: `Link`, // dangerous data URL removes the link
		},
		{
			name:     "Meta refresh with javascript",
			input:    `<meta http-equiv="refresh" content="0;url=javascript:alert('xss')">`,
			expected: ``,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeHTML(tc.input)
			if result != tc.expected {
				t.Errorf("sanitizeHTML(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Script tag in text",
			input:    `<script>alert('xss')</script>`,
			expected: ``,
		},
		{
			name:     "HTML tags in text",
			input:    `<p>Hello <strong>World</strong>!</p>`,
			expected: `Hello World!`,
		},
		{
			name:     "Plain text",
			input:    `Hello World! This is plain text.`,
			expected: `Hello World! This is plain text.`,
		},
		{
			name:     "Text with special characters",
			input:    `Hello & goodbye < > " '`,
			expected: `Hello &amp; goodbye &lt; &gt; &#34; &#39;`,
		},
		{
			name:     "Image tag in text",
			input:    `<img src=x onerror=alert('xss')>`,
			expected: ``,
		},
		{
			name:     "Mixed HTML and text",
			input:    `Normal text <script>evil()</script> more text`,
			expected: `Normal text  more text`,
		},
		{
			name:     "Unicode characters",
			input:    `Hello 世界 🌍`,
			expected: `Hello 世界 🌍`,
		},
		{
			name:     "Multiple script tags",
			input:    `<script>alert(1)</script>Text<script>alert(2)</script>`,
			expected: `Text`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeText(tc.input)
			if result != tc.expected {
				t.Errorf("sanitizeText(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestBuildUser_XSS(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedEmail string
		expectedName  string
		shouldBeNil   bool
	}{
		{
			name:          "Normal user",
			input:         `John Doe <john@example.com>`,
			expectedEmail: "john@example.com",
			expectedName:  "John Doe",
			shouldBeNil:   false,
		},
		{
			name:          "Simple email only",
			input:         `test@example.com`,
			expectedEmail: "test@example.com",
			expectedName:  "",
			shouldBeNil:   false,
		},
		{
			name:          "User with quotes",
			input:         `"John O'Connor" <john@example.com>`,
			expectedEmail: "john@example.com",
			expectedName:  "John O&#39;Connor",
			shouldBeNil:   false,
		},
		{
			name:          "JavaScript in name - should be filtered",
			input:         `javascript:alert('xss') <js@example.com>`,
			expectedEmail: "js@example.com",
			expectedName:  "javascript:alert(&#39;xss&#39;)",
			shouldBeNil:   false,
		},
		{
			name:          "Event handler in name - quotes removed in sanitization",
			input:         `onclick="alert('xss')" <click@example.com>`,
			expectedEmail: "click@example.com",
			expectedName:  "onclick= alert(&#39;xss&#39;)", 
			shouldBeNil:   false,
		},
		{
			name:          "HTML in email address - should be filtered",
			input:         `<script>test@example.com</script>`,
			expectedEmail: "script>test@example.com</script",
			expectedName:  "",
			shouldBeNil:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := buildUser(tc.input)
			
			if tc.shouldBeNil {
				if user != nil {
					t.Errorf("buildUser(%q) should return nil due to heavy sanitization, but got: %+v", tc.input, user)
				}
				return
			}
			
			if user == nil {
				t.Fatalf("buildUser returned nil for input: %q", tc.input)
			}
			if user.EmailAddress != tc.expectedEmail {
				t.Errorf("buildUser(%q).EmailAddress = %q, want %q", tc.input, user.EmailAddress, tc.expectedEmail)
			}
			if user.Name != tc.expectedName {
				t.Errorf("buildUser(%q).Name = %q, want %q", tc.input, user.Name, tc.expectedName)
			}
		})
	}
}

func TestNewEmailFromReader_XSS(t *testing.T) {
	var b bytes.Buffer

	var h message.Header
	h.Set("From", `Safe Name <from@example.com>`)
	h.Set("To", "to@example.com")
	h.Set("Subject", `<script>alert('xss')</script>Test Subject`)
	h.Set("Date", "Mon, 02 Jan 2006 15:04:05 -0700")
	h.SetContentType("multipart/alternative", nil)

	w, err := message.CreateWriter(&b, h)
	if err != nil {
		t.Fatal(err)
	}

	var htmlHeader message.Header
	htmlHeader.SetContentType("text/html", nil)
	htmlPart, err := w.CreatePart(htmlHeader)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(htmlPart, `<h1>Hello</h1><script>alert('xss')</script><p>Safe content</p>`)
	htmlPart.Close()

	var textHeader message.Header
	textHeader.SetContentType("text/plain", nil)
	textPart, err := w.CreatePart(textHeader)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(textPart, `Hello <script>alert('xss')</script> World`)
	textPart.Close()

	w.Close()

	email := NewEmailFromReader([]string{"to@example.com"}, &b, b.Len())

	if email.From.EmailAddress != "from@example.com" {
		t.Errorf("Expected From.EmailAddress to be 'from@example.com', got: %q", email.From.EmailAddress)
	}

	if email.From.Name != "Safe Name" {
		t.Errorf("Expected From.Name to be 'Safe Name', got: %q", email.From.Name)
	}

	expectedSubject := "Test Subject"
	if email.Subject != expectedSubject {
		t.Errorf("Expected Subject to be %q, got: %q", expectedSubject, email.Subject)
	}

	htmlContent := string(email.HTML)
	if strings.Contains(htmlContent, "<script>") {
		t.Errorf("HTML content should not contain script tags, got: %q", htmlContent)
	}
	if !strings.Contains(htmlContent, "Hello") {
		t.Errorf("HTML content should contain 'Hello', got: %q", htmlContent)
	}

	textContent := string(email.Text)
	if strings.Contains(textContent, "<script>") {
		t.Errorf("Text content should not contain any HTML tags, got: %q", textContent)
	}
	if !strings.Contains(textContent, "Hello") || !strings.Contains(textContent, "World") {
		t.Errorf("Text content should keep safe text, got: %q", textContent)
	}
}

func TestAttachment_XSS(t *testing.T) {
	var b bytes.Buffer

	var h message.Header
	h.Set("From", "test@example.com")
	h.Set("To", "to@example.com")
	h.Set("Subject", "Test")
	h.SetContentType("multipart/mixed", nil)

	w, err := message.CreateWriter(&b, h)
	if err != nil {
		t.Fatal(err)
	}

	var attachHeader message.Header
	attachHeader.Set("Content-Type", "application/octet-stream")
	attachHeader.Set("Content-Disposition", `attachment; filename="test.txt"`)
	attachPart, err := w.CreatePart(attachHeader)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(attachPart, "test content")
	attachPart.Close()

	w.Close()

	email := NewEmailFromReader([]string{"to@example.com"}, &b, b.Len())

	if len(email.Attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(email.Attachments))
	}

	// Test that sanitizeText function works correctly on attachment data
	dangerousFilename := `<script>alert('xss')</script>malicious.txt`
	sanitizedFilename := sanitizeText(dangerousFilename)
	if strings.Contains(sanitizedFilename, "<script>") {
		t.Errorf("Sanitized filename should not contain script tags, got: %q", sanitizedFilename)
	}

	dangerousContentType := `application/octet-stream<script>alert('xss')</script>`
	sanitizedContentType := sanitizeText(dangerousContentType)
	if strings.Contains(sanitizedContentType, "<script>") {
		t.Errorf("Sanitized content type should not contain script tags, got: %q", sanitizedContentType)
	}
}

func TestComplexXSSScenarios(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		testType string // "html" or "text"
	}{
		{
			name:     "Nested script tags",
			input:    `<script><script>alert('xss')</script></script>`,
			testType: "html",
		},
		{
			name:     "Base64 encoded script",
			input:    `<img src="data:image/svg+xml;base64,PHNjcmlwdD5hbGVydCgneHNzJyk8L3NjcmlwdD4=">`,
			testType: "html",
		},
		{
			name:     "CSS expression - should be removed",
			input:    `<div style="background:expression(alert('xss'))">Test</div>`,
			testType: "html",
		},
		{
			name:     "HTML entities script",
			input:    `&lt;script&gt;alert('xss')&lt;/script&gt;`,
			testType: "html",
		},
		{
			name:     "Multiple attack vectors",
			input:    `<script>alert(1)</script><img src=x onerror=alert(2)><svg onload=alert(3)>`,
			testType: "html",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result string
			if tc.testType == "html" {
				result = sanitizeHTML(tc.input)
			} else {
				result = sanitizeText(tc.input)
			}

			// 检查结果中不应包含危险的内容
			dangerousPatterns := []string{
				"<script",
				"javascript:",
				"onerror=",
				"onload=",
				"onclick=",
				"expression(",
			}

			for _, pattern := range dangerousPatterns {
				if strings.Contains(strings.ToLower(result), pattern) {
					t.Errorf("Sanitized result contains dangerous pattern %q: %q", pattern, result)
				}
			}
		})
	}
}

func TestXSSInMultipleFields(t *testing.T) {
	var b bytes.Buffer

	var h message.Header
	h.Set("From", `Normal Name <from@example.com>`)
	h.Set("To", "to@example.com")
	h.Set("Subject", `<script>alert('subject')</script>Important Email`)
	h.Set("Date", "Mon, 02 Jan 2006 15:04:05 -0700")
	h.SetContentType("text/plain", nil)

	w, err := message.CreateWriter(&b, h)
	if err != nil {
		t.Fatal(err)
	}

	io.WriteString(w, `Email body with <script>alert('body')</script> content`)
	w.Close()

	email := NewEmailFromReader([]string{"to@example.com"}, &b, b.Len())

	if strings.Contains(email.Subject, "<script>") {
		t.Errorf("Subject contains unsanitized script tag: %q", email.Subject)
	}

	textContent := string(email.Text)
	if strings.Contains(textContent, "<script>") {
		t.Errorf("Text content contains unsanitized script tag: %q", textContent)
	}

	if !strings.Contains(textContent, "Email body") || !strings.Contains(textContent, "content") {
		t.Errorf("Text content should keep safe text, got: %q", textContent)
	}

	if !strings.Contains(email.Subject, "Important Email") {
		t.Errorf("Subject should keep safe text, got: %q", email.Subject)
	}
} 