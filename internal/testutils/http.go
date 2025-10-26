package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HTTPTestSuite contains common utilities for HTTP testing
type HTTPTestSuite struct {
	Router *gin.Engine
}

// SetupHTTPTest initializes Gin for testing
func SetupHTTPTest() *HTTPTestSuite {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	return &HTTPTestSuite{
		Router: router,
	}
}

// MakeRequest creates and executes an HTTP request for testing
func (suite *HTTPTestSuite) MakeRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBody io.Reader

	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	suite.Router.ServeHTTP(recorder, req)

	return recorder
}

// MakeRequestWithHeaders creates and executes an HTTP request with custom headers
func (suite *HTTPTestSuite) MakeRequestWithHeaders(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody io.Reader

	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	suite.Router.ServeHTTP(recorder, req)

	return recorder
}

// AssertJSONResponse asserts the response status and unmarshals JSON response
func AssertJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	assert.Equal(t, expectedStatus, recorder.Code)
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

	if target != nil {
		err := json.Unmarshal(recorder.Body.Bytes(), target)
		require.NoError(t, err)
	}
}

// AssertErrorResponse asserts an error response with specific message
func AssertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	assert.Equal(t, expectedStatus, recorder.Code)

	var errorResponse map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	if expectedMessage != "" {
		assert.Contains(t, errorResponse["error"], expectedMessage)
	}
}

// AssertSuccessResponse asserts a successful response
func AssertSuccessResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(t, expectedStatus, recorder.Code)
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
}

// ParseJSONResponse parses JSON response into target struct
func ParseJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(recorder.Body.Bytes(), target)
	require.NoError(t, err)
}

// MockHTTPRequest represents a mock HTTP request for testing
type MockHTTPRequest struct {
	Method  string
	URL     string
	Body    interface{}
	Headers map[string]string
}

// MockHTTPResponse represents a mock HTTP response for testing
type MockHTTPResponse struct {
	Status int
	Body   interface{}
}

// HTTPTestCase represents a test case for HTTP handlers
type HTTPTestCase struct {
	Name             string
	Request          MockHTTPRequest
	ExpectedResponse MockHTTPResponse
	Setup            func()
	Teardown         func()
}

// RunHTTPTestCases runs a series of HTTP test cases
func (suite *HTTPTestSuite) RunHTTPTestCases(t *testing.T, testCases []HTTPTestCase) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Setup != nil {
				tc.Setup()
			}

			var recorder *httptest.ResponseRecorder
			if len(tc.Request.Headers) > 0 {
				recorder = suite.MakeRequestWithHeaders(
					tc.Request.Method,
					tc.Request.URL,
					tc.Request.Body,
					tc.Request.Headers,
				)
			} else {
				recorder = suite.MakeRequest(
					tc.Request.Method,
					tc.Request.URL,
					tc.Request.Body,
				)
			}

			// Assert status code
			assert.Equal(t, tc.ExpectedResponse.Status, recorder.Code)

			// Assert response body if provided
			if tc.ExpectedResponse.Body != nil {
				var actualResponse interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
				require.NoError(t, err)

				expectedJSON, _ := json.Marshal(tc.ExpectedResponse.Body)
				actualJSON, _ := json.Marshal(actualResponse)
				assert.JSONEq(t, string(expectedJSON), string(actualJSON))
			}

			if tc.Teardown != nil {
				tc.Teardown()
			}
		})
	}
}

// CreateTestGinContext creates a test Gin context
func CreateTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	return ctx, recorder
}

// SetJSONBody sets JSON body for a Gin context
func SetJSONBody(ctx *gin.Context, body interface{}) {
	jsonBytes, _ := json.Marshal(body)
	ctx.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBytes))
	ctx.Request.Header.Set("Content-Type", "application/json")
}

// SetURLParam sets URL parameter for a Gin context
func SetURLParam(ctx *gin.Context, key, value string) {
	ctx.Params = gin.Params{
		{Key: key, Value: value},
	}
}

// SetQueryParam sets query parameter for a Gin context
func SetQueryParam(ctx *gin.Context, key, value string) {
	if ctx.Request.URL == nil {
		ctx.Request.URL = &url.URL{}
	}
	if ctx.Request.URL.RawQuery == "" {
		ctx.Request.URL.RawQuery = key + "=" + value
	} else {
		ctx.Request.URL.RawQuery += "&" + key + "=" + value
	}
}
