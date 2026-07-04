package parser_test

import (
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/parser"
	"github.com/stretchr/testify/assert"
)

func Test_ParseLog(t *testing.T) {
	type TestCase struct {
		name     string
		input    string
		expected parser.LogInfo
	}

	TestCases := []TestCase{
		{
			name: "Пустой remote_user, GET запрос, пустой http referer",
			input: "93.180.71.3 - - [17/May/2015:08:05:57 +0000] \"GET /downloads/product_1 HTTP/1.1\" 304 " +
				"0 \"-\" \"Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)\"",
			expected: parser.LogInfo{
				RemoteAddr:    "93.180.71.3",
				RemoteUser:    "-",
				TimeLocal:     "17/May/2015:08:05:57 +0000",
				Method:        "GET",
				RequestURL:    "/downloads/product_1",
				HTTPVersion:   "HTTP/1.1",
				Status:        "304",
				BodyBytesSent: "0",
				HTTPReferer:   "-",
				HTTPUserAgent: "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
			},
		},
		{
			name: "Неустой remote_user, PATCH запрос, пустой http referer",
			input: "216.46.173.126 - RemoteUsr [17/May/2015:08:05:57 +0000] \"PATCH /downloads/product_1 HTTP/1.1\" 304 " +
				"0 \"-\" \"Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)\"",
			expected: parser.LogInfo{
				RemoteAddr:    "216.46.173.126",
				RemoteUser:    "RemoteUsr",
				TimeLocal:     "17/May/2015:08:05:57 +0000",
				Method:        "PATCH",
				RequestURL:    "/downloads/product_1",
				HTTPVersion:   "HTTP/1.1",
				Status:        "304",
				BodyBytesSent: "0",
				HTTPReferer:   "-",
				HTTPUserAgent: "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
			},
		},
		{
			name: "Непустой remote_user, PUT запрос, непустой http referer, статус 200",
			input: "0.0.0.0 - ABCDE [17/May/2015:08:05:57 +0000] \"PUT /downloads/product_1 HTTP/1.1\" 200 " +
				"0 \"qwerty\" \"Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)\"",
			expected: parser.LogInfo{
				RemoteAddr:    "0.0.0.0",
				RemoteUser:    "ABCDE",
				TimeLocal:     "17/May/2015:08:05:57 +0000",
				Method:        "PUT",
				RequestURL:    "/downloads/product_1",
				HTTPVersion:   "HTTP/1.1",
				Status:        "200",
				BodyBytesSent: "0",
				HTTPReferer:   "qwerty",
				HTTPUserAgent: "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
			},
		},
		{
			name: "Пустой remote_user, DELETE запрос, пустой http referer",
			input: "127.0.0.1 - - [26/Jul/2004:08:05:57 +0000] \"DELETE /downloads/product_1 HTTP/1.1\" 304 " +
				"0 \"-\" \"ioi\"",
			expected: parser.LogInfo{
				RemoteAddr:    "127.0.0.1",
				RemoteUser:    "-",
				TimeLocal:     "26/Jul/2004:08:05:57 +0000",
				Method:        "DELETE",
				RequestURL:    "/downloads/product_1",
				HTTPVersion:   "HTTP/1.1",
				Status:        "304",
				BodyBytesSent: "0",
				HTTPReferer:   "-",
				HTTPUserAgent: "ioi",
			},
		},
	}

	for _, tc := range TestCases {
		assert.Equal(t, tc.expected, *parser.ParseLogString(tc.input), tc.name)
	}
}
