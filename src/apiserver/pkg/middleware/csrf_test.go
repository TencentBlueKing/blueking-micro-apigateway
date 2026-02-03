/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
 * Copyright (C) 2025 Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 *     http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */
package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	csrf "filippo.io/csrf/gorilla"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCSRF_GETRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// GET requests are safe methods and should be allowed
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRF_POSTRequestWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// POST requests without CSRF token should be rejected
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRF_POSTRequestWithValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	// Create a router with both CSRF protection and token setting
	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.Use(CSRFToken(appID, ""))
	router.GET("/token", func(c *gin.Context) {
		token := csrf.Token(c.Request)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a test server
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Create a client with cookie jar to handle cookies properly
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{
		Jar: jar,
	}

	// First, make a GET request to get the CSRF token
	tokenResp, err := client.Get(ts.URL + "/token")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, tokenResp.StatusCode)
	defer tokenResp.Body.Close()

	// Verify that the CSRF protection cookie is set
	tsURL, err := url.Parse(ts.URL)
	assert.NoError(t, err)
	cookies := jar.Cookies(tsURL)
	csrfCookieFound := false
	for _, cookie := range cookies {
		if cookie.Name == appID+"-csrf" {
			csrfCookieFound = true
			assert.NotEmpty(t, cookie.Value, "CSRF cookie should have a value")
			break
		}
	}
	assert.True(t, csrfCookieFound, "CSRF protection cookie should be set")

	// Parse the token from JSON response
	var tokenResponse map[string]string
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	assert.NoError(t, err)
	csrfTokenValue, ok := tokenResponse["token"]
	assert.True(t, ok, "Token should be present in response")
	assert.NotEmpty(t, csrfTokenValue, "Token should not be empty")

	// Make a POST request with the CSRF token in the header
	// The cookie jar will automatically include the cookies
	postReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/test", strings.NewReader("data=test"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Set the token in the header (gorilla/csrf accepts X-CSRF-Token header, case-insensitive)
	postReq.Header.Set("X-CSRF-Token", csrfTokenValue)

	postResp, err := client.Do(postReq)
	assert.NoError(t, err)
	defer postResp.Body.Close()

	// POST request with valid CSRF token should be allowed
	assert.Equal(t, http.StatusOK, postResp.StatusCode)
}

func TestCSRFToken_SetsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	domain := "example.com"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	// CSRFToken requires CSRF middleware to be applied first
	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.Use(CSRFToken(appID, domain))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the CSRF token cookie is set
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == appID+"-csrf-token" {
			assert.NotEmpty(t, cookie.Value)
			assert.Equal(t, "/", cookie.Path)
			assert.Equal(t, domain, cookie.Domain)
			assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
			found = true
			break
		}
	}
	assert.True(t, found, "CSRF token cookie should be set")
}

func TestCSRF_CookieName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "my-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the CSRF cookie has the correct name
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == appID+"-csrf" {
			found = true
			break
		}
	}
	assert.True(t, found, "CSRF cookie should have the correct name with appID prefix")
}

// TestCSRF_PlaintextHTTPRequest 测试 PlaintextHTTPRequest 的使用
func TestCSRF_PlaintextHTTPRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试 POST 请求在非 HTTPS 环境下的行为
	// 由于使用了 PlaintextHTTPRequest，应该跳过 Referer 检查
	req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader("data=test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 不设置 Referer 头，在 v1.7.3 之前这会导致失败，但现在应该通过

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 403，因为没有 CSRF token，但不是因为 Referer 检查失败
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestCSRF_PlaintextHTTPRequestWithToken 测试带有有效 token 的 PlaintextHTTPRequest
func TestCSRF_PlaintextHTTPRequestWithToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.Use(CSRFToken(appID, ""))
	router.GET("/token", func(c *gin.Context) {
		token := csrf.Token(c.Request)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 创建测试服务器
	ts := httptest.NewServer(router)
	defer ts.Close()

	// 创建带 cookie jar 的客户端
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}

	// 首先获取 CSRF token
	tokenResp, err := client.Get(ts.URL + "/token")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, tokenResp.StatusCode)
	defer tokenResp.Body.Close()

	// 解析 token
	var tokenResponse map[string]string
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	assert.NoError(t, err)
	csrfTokenValue := tokenResponse["token"]
	assert.NotEmpty(t, csrfTokenValue)

	// 发送 POST 请求，不设置 Referer 头但包含有效的 CSRF token
	postReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/test", strings.NewReader("data=test"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.Header.Set("X-CSRF-Token", csrfTokenValue)
	// 故意不设置 Referer 头来测试 PlaintextHTTPRequest 的效果

	postResp, err := client.Do(postReq)
	assert.NoError(t, err)
	defer postResp.Body.Close()

	// 应该成功，因为 PlaintextHTTPRequest 跳过了 Referer 检查
	assert.Equal(t, http.StatusOK, postResp.StatusCode)
}

// TestCSRF_ContextKeyPresence 测试 PlaintextHTTPRequest 是否正确设置了上下文键
func TestCSRF_ContextKeyPresence(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	var contextChecked bool
	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.GET("/test", func(c *gin.Context) {
		// 检查请求上下文中是否设置了 PlaintextHTTPContextKey
		if c.Request.Context().Value(csrf.PlaintextHTTPContextKey) != nil {
			contextChecked = true
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, contextChecked, "PlaintextHTTPContextKey should be set in request context")
}

// TestCSRF_MiddlewareStructure 测试中间件的结构和调用顺序
func TestCSRF_MiddlewareStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	var middlewareOrder []string
	router := gin.New()

	// 添加一个前置中间件来记录调用顺序
	router.Use(func(c *gin.Context) {
		middlewareOrder = append(middlewareOrder, "before-csrf")
		c.Next()
	})

	router.Use(CSRF(appID, secret, trustedOrigins))

	// 添加一个后置中间件
	router.Use(func(c *gin.Context) {
		middlewareOrder = append(middlewareOrder, "after-csrf")
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		middlewareOrder = append(middlewareOrder, "handler")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedOrder := []string{"before-csrf", "after-csrf", "handler"}
	assert.Equal(t, expectedOrder, middlewareOrder, "Middleware should be called in correct order")
}

// TestCSRF_ErrorHandling 测试 CSRF 错误处理
func TestCSRF_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{}

	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试各种无效的 CSRF token 情况
	testCases := []struct {
		name        string
		tokenHeader string
		tokenForm   string
		expectCode  int
	}{
		{
			name:       "no token",
			expectCode: http.StatusForbidden,
		},
		{
			name:        "invalid token in header",
			tokenHeader: "invalid-token",
			expectCode:  http.StatusForbidden,
		},
		{
			name:       "invalid token in form",
			tokenForm:  "invalid-token",
			expectCode: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body strings.Builder
			body.WriteString("data=test")
			if tc.tokenForm != "" {
				body.WriteString("&gorilla.csrf.Token=" + tc.tokenForm)
			}

			req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(body.String()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			if tc.tokenHeader != "" {
				req.Header.Set("X-CSRF-Token", tc.tokenHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectCode, w.Code, "Expected status code for %s", tc.name)
		})
	}
}

// TestCSRF_TrustedOrigins 测试 TrustedOrigins 配置
func TestCSRF_TrustedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{"example.com", "trusted.example.org"}

	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.Use(CSRFToken(appID, ""))
	router.GET("/token", func(c *gin.Context) {
		token := csrf.Token(c.Request)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 创建测试服务器
	ts := httptest.NewServer(router)
	defer ts.Close()

	// 创建带 cookie jar 的客户端
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}

	// 首先获取 CSRF token
	tokenResp, err := client.Get(ts.URL + "/token")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, tokenResp.StatusCode)
	defer tokenResp.Body.Close()

	// 解析 token
	var tokenResponse map[string]string
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	assert.NoError(t, err)
	csrfTokenValue := tokenResponse["token"]
	assert.NotEmpty(t, csrfTokenValue)

	// 发送 POST 请求，设置来自可信源的 Origin 头
	postReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/test", strings.NewReader("data=test"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.Header.Set("X-CSRF-Token", csrfTokenValue)
	postReq.Header.Set("Origin", "https://example.com")

	postResp, err := client.Do(postReq)
	assert.NoError(t, err)
	defer postResp.Body.Close()

	// 应该成功，因为 Origin 在 TrustedOrigins 列表中
	assert.Equal(t, http.StatusOK, postResp.StatusCode)
}

// TestCSRF_UntrustedOrigin 测试不可信源的 Origin 被拒绝
func TestCSRF_UntrustedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"
	trustedOrigins := []string{"example.com"}

	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.Use(CSRFToken(appID, ""))
	router.GET("/token", func(c *gin.Context) {
		token := csrf.Token(c.Request)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 创建测试服务器
	ts := httptest.NewServer(router)
	defer ts.Close()

	// 创建带 cookie jar 的客户端
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}

	// 首先获取 CSRF token
	tokenResp, err := client.Get(ts.URL + "/token")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, tokenResp.StatusCode)
	defer tokenResp.Body.Close()

	// 解析 token
	var tokenResponse map[string]string
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	assert.NoError(t, err)
	csrfTokenValue := tokenResponse["token"]
	assert.NotEmpty(t, csrfTokenValue)

	// 发送 POST 请求，设置来自不可信源的 Origin 头
	postReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/test", strings.NewReader("data=test"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.Header.Set("X-CSRF-Token", csrfTokenValue)
	postReq.Header.Set("Origin", "https://malicious.com")

	postResp, err := client.Do(postReq)
	assert.NoError(t, err)
	defer postResp.Body.Close()

	// 应该返回 403 Forbidden，因为 Origin 不在 TrustedOrigins 列表中
	assert.Equal(t, http.StatusForbidden, postResp.StatusCode)
}
