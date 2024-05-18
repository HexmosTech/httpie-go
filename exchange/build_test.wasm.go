//go:build wasm

package exchange

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/HexmosTech/httpie-go/input"
)

func parseURL(t *testing.T, rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		t.Fatalf("failed to parse URL: %s", err)
	}
	return u
}


func TestBuildURL(t *testing.T) {
	testCases := []struct {
		title      string
		url        string
		parameters []input.Field
		expected   string
	}{
		{
			title: "Typical case",
			url:   "http://example.com/hello",
			parameters: []input.Field{
				{Name: "foo", Value: "bar"},
				{Name: "fizz", Value: "buzz"},
			},
			expected: "http://example.com/hello?fizz=buzz&foo=bar",
		},
		{
			title: "Both URL and Parameters have query string",
			url:   "http://example.com/hello?hoge=fuga",
			parameters: []input.Field{
				{Name: "foo", Value: "bar"},
				{Name: "fizz", Value: "buzz"},
			},
			expected: "http://example.com/hello?fizz=buzz&foo=bar&hoge=fuga",
		},
		{
			title: "Multiple values with a key",
			url:   "http://example.com/hello",
			parameters: []input.Field{
				{Name: "foo", Value: "value 1"},
				{Name: "foo", Value: "value 2"},
				{Name: "foo", Value: "value 3"},
			},
			expected: "http://example.com/hello?foo=value+1&foo=value+2&foo=value+3",
		},
		{
			title: "Multiple values with a key in both URL and Parameters",
			url:   "http://example.com/hello?foo=a&foo=z",
			parameters: []input.Field{
				{Name: "foo", Value: "value 1"},
				{Name: "foo", Value: "value 2"},
				{Name: "foo", Value: "value 3"},
			},
			expected: "http://example.com/hello?foo=a&foo=z&foo=value+1&foo=value+2&foo=value+3",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			in := &input.Input{
				URL:        parseURL(t, tt.url),
				Parameters: tt.parameters,
			}
			u, err := buildURL(in)
			if err != nil {
				t.Fatalf("unexpected error: err=%v", err)
			}
			if u.String() != tt.expected {
				t.Errorf("unexpected URL: expected=%s, actual=%s", tt.expected, u)
			}
		})
	}
}

func makeTempFile(t *testing.T, content string) string {
	tmpfile, err := ioutil.TempFile("", "httpie-go-test-")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		os.Remove(tmpfile.Name())
		t.Fatalf("failed to write to temporary file: %v", err)
	}
	return tmpfile.Name()
}

func TestBuildHTTPHeader(t *testing.T) {
	// Setup
	fileName := makeTempFile(t, "test test")
	defer os.Remove(fileName)
	header := input.Header{
		Fields: []input.Field{
			{Name: "X-Foo", Value: "foo", IsFile: false},
			{Name: "X-From-File", Value: fileName, IsFile: true},
			{Name: "X-Multi-Value", Value: "value 1"},
			{Name: "X-Multi-Value", Value: "value 2"},
		},
	}
	in := &input.Input{Header: header}

	// Exercise
	httpHeader, err := buildHTTPHeader(in)
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	// Verify
	expected := http.Header{
		"X-Foo":         []string{"foo"},
		"X-From-File":   []string{"test test"},
		"X-Multi-Value": []string{"value 1", "value 2"},
	}
	if !reflect.DeepEqual(httpHeader, expected) {
		t.Errorf("unexpected header: expected=%v, actual=%v", expected, httpHeader)
	}
}

func isEquivalentJSON(t *testing.T, json1, json2 string) bool {
	var obj1, obj2 interface{}
	if err := json.Unmarshal([]byte(json1), &obj1); err != nil {
		t.Fatalf("failed to unmarshal json1: %v", err)
	}
	if err := json.Unmarshal([]byte(json2), &obj2); err != nil {
		t.Fatalf("failed to unmarshal json2: %v", err)
	}
	return reflect.DeepEqual(obj1, obj2)
}

func readAll(t *testing.T, reader io.Reader) string {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read all: %s", err)
	}
	return string(b)
}

func TestBuildHTTPBody_EmptyBody(t *testing.T) {
	// Setup
	fileName := makeTempFile(t, "test test")
	defer os.Remove(fileName)
	body := input.Body{
		BodyType: input.EmptyBody,
	}
	in := &input.Input{Body: body}

	// Exercise
	actual, err := buildHTTPBody(in)
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	// Verify
	expected := bodyTuple{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected body tuple: expected=%+v, actual=%+v", expected, actual)
	}
}

func TestBuildHTTPBody_JSONBody(t *testing.T) {
	// Setup
	fileName := makeTempFile(t, "test test")
	defer os.Remove(fileName)
	body := input.Body{
		BodyType: input.JSONBody,
		Fields: []input.Field{
			{Name: "foo", Value: "bar"},
			{Name: "from_file", Value: fileName, IsFile: true},
		},
		RawJSONFields: []input.Field{
			{Name: "boolean", Value: "true"},
			{Name: "array", Value: `[1, null, "hello"]`},
		},
	}
	in := &input.Input{Body: body}

	// Exercise
	bodyTuple, err := buildHTTPBody(in)
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	// Verify
	expectedBody := `{
		"foo": "bar",
		"from_file": "test test",
		"boolean": true,
		"array": [1, null, "hello"]
	}`
	actualBody := readAll(t, bodyTuple.body)
	if !isEquivalentJSON(t, expectedBody, actualBody) {
		t.Errorf("unexpected body: expected=%s, actual=%s", expectedBody, actualBody)
	}
	expectedContentType := "application/json"
	if bodyTuple.contentType != expectedContentType {
		t.Errorf("unexpected content type: expected=%s, actual=%s", expectedContentType, bodyTuple.contentType)
	}
	if bodyTuple.contentLength != int64(len(actualBody)) {
		t.Errorf("invalid content length: len(body)=%v, actual=%v", len(actualBody), bodyTuple.contentLength)
	}
}

func TestBuildHTTPBody_FormBody_URLEncoded(t *testing.T) {
	// Setup
	fileName := makeTempFile(t, "love & peace")
	defer os.Remove(fileName)
	body := input.Body{
		BodyType: input.FormBody,
		Fields: []input.Field{
			{Name: "foo", Value: "bar"},
			{Name: "from_file", Value: fileName, IsFile: true},
		},
	}
	in := &input.Input{Body: body}

	// Exercise
	bodyTuple, err := buildHTTPBody(in)
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	// Verify
	expectedBody := `foo=bar&from_file=love+%26+peace`
	actualBody := readAll(t, bodyTuple.body)
	if actualBody != expectedBody {
		t.Errorf("unexpected body: expected=%s, actual=%s", expectedBody, actualBody)
	}
	expectedContentType := "application/x-www-form-urlencoded; charset=utf-8"
	if bodyTuple.contentType != expectedContentType {
		t.Errorf("unexpected content type: expected=%s, actual=%s", expectedContentType, bodyTuple.contentType)
	}
	if bodyTuple.contentLength != int64(len(actualBody)) {
		t.Errorf("invalid content length: len(body)=%v, actual=%v", len(actualBody), bodyTuple.contentLength)
	}
}

func TestBuildHTTPBody_FormBody_Multipart(t *testing.T) {
	// Setup
	fileName := makeTempFile(t, "🍣 & 🍺")
	defer os.Remove(fileName)
	body := input.Body{
		BodyType: input.FormBody,
		Fields: []input.Field{
			{Name: "hello", Value: "🍺 world!"},
			{Name: `"double-quoted"`, Value: "should be escaped"},
		},
		Files: []input.Field{
			{Name: "file1", Value: fileName, IsFile: true},
			{Name: "file2", Value: "From STDIN", IsFile: false},
		},
	}
	in := &input.Input{Body: body}

	// Exercise
	bodyTuple, err := buildHTTPBody(in)
	if err != nil {
		t.Fatalf("unexpected error: err=%+v", err)
	}

	// Verify
	expectedBody := regexp.MustCompile(strings.Join([]string{
		`--[0-9a-z]+`,
		regexp.QuoteMeta(`Content-Disposition: form-data; name="hello"`),
		regexp.QuoteMeta(``),
		regexp.QuoteMeta(`🍺 world!`),
		`--[0-9a-z]+`,
		regexp.QuoteMeta(`Content-Disposition: form-data; name*=utf-8''%22double-quoted%22`),
		regexp.QuoteMeta(``),
		regexp.QuoteMeta(`should be escaped`),
		`--[0-9a-z]+`,
		regexp.QuoteMeta(`Content-Disposition: form-data; name="file1"; filename="` + path.Base(fileName) + `"`),
		regexp.QuoteMeta(``),
		regexp.QuoteMeta(`🍣 & 🍺`),
		`--[0-9a-z]+`,
		regexp.QuoteMeta(`Content-Disposition: form-data; name="file2"`),
		regexp.QuoteMeta(``),
		regexp.QuoteMeta(`From STDIN`),
		`--[0-9a-z]+--`,
		regexp.QuoteMeta(``),
	}, "\r\n"))

	actualBody := readAll(t, bodyTuple.body)
	if !expectedBody.MatchString(actualBody) {
		t.Errorf("unexpected body: expected='%s', actual='%s'", expectedBody, actualBody)
	}
	expectedContentType := "multipart/form-data; "
	if !strings.HasPrefix(bodyTuple.contentType, expectedContentType) {
		t.Errorf("unexpected content type: expected=%s, actual=%s", expectedContentType, bodyTuple.contentType)
	}
	if bodyTuple.contentLength != int64(len(actualBody)) {
		t.Errorf("invalid content length: len(body)=%v, actual=%v", len(actualBody), bodyTuple.contentLength)
	}
}
