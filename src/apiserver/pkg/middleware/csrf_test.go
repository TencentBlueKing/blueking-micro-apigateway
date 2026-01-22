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

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
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

// TestCSRF_RealWorldScenario 测试真实场景：模拟从环境变量配置的 Origin 进行 API 访问
func TestCSRF_RealWorldScenario(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "bk-micro-gateway"
	secret := "test-secret-key-32-bytes-long!!"

	// 模拟真实配置：从环境变量读取的完整 URL 经过 ExtractHostsForCSRF 处理后得到的主机名列表
	// 原始配置: ALLOWED_ORIGINS=https://dev-t.paas3-dev.bktencent.com:8888,https://bk-micro-web.paas3-dev.bktencent.com
	// 处理后的 CSRF trusted origins:
	trustedOrigins := []string{
		"dev-t.paas3-dev.bktencent.com:8888",
		"bk-micro-web.paas3-dev.bktencent.com",
	}

	router := gin.New()
	router.Use(CSRF(appID, secret, trustedOrigins))
	router.Use(CSRFToken(appID, ""))
	router.GET("/api/v1/web/token", func(c *gin.Context) {
		token := csrf.Token(c.Request)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	router.PUT("/api/v1/web/gateways/:gateway_id/routes/:id/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"gateway_id": c.Param("gateway_id"),
			"route_id":   c.Param("id"),
		})
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}

	// Step 1: 获取 CSRF token
	tokenResp, err := client.Get(ts.URL + "/api/v1/web/token")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, tokenResp.StatusCode)
	defer tokenResp.Body.Close()

	var tokenResponse map[string]string
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	assert.NoError(t, err)
	csrfToken := tokenResponse["token"]
	assert.NotEmpty(t, csrfToken)

	// Step 2: 测试来自可信源的 PUT 请求（模拟编辑路由）
	testCases := []struct {
		name       string
		origin     string
		expectCode int
	}{
		{
			name:       "trusted origin without port",
			origin:     "https://bk-micro-web.paas3-dev.bktencent.com",
			expectCode: http.StatusOK,
		},
		{
			name:       "trusted origin with port",
			origin:     "https://dev-t.paas3-dev.bktencent.com:8888",
			expectCode: http.StatusOK,
		},
		{
			name:       "untrusted origin",
			origin:     "https://malicious.com",
			expectCode: http.StatusForbidden,
		},
		{
			name:       "similar but different origin",
			origin:     "https://fake-bk-micro-web.paas3-dev.bktencent.com",
			expectCode: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟真实请求：PUT /api/v1/web/gateways/26/routes/bk.r.xxx/
			reqBody := `{"config":{"methods":["GET","POST"],"uris":["aaaa"],"name":"route33"}}`
			putReq, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/web/gateways/26/routes/bk.r.hQkvuSAAOG/", strings.NewReader(reqBody))
			putReq.Header.Set("Content-Type", "application/json")
			putReq.Header.Set("X-CSRF-Token", csrfToken)
			putReq.Header.Set("Origin", tc.origin)
			putReq.Header.Set("X-Requested-With", "fetch")

			putResp, err := client.Do(putReq)
			assert.NoError(t, err)
			defer putResp.Body.Close()

			assert.Equal(t, tc.expectCode, putResp.StatusCode,
				"Origin %s should return %d", tc.origin, tc.expectCode)
		})
	}
}

// TestCSRF_CORSAndCSRFIntegration 测试 CORS 和 CSRF 中间件的集成
func TestCSRF_CORSAndCSRFIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appID := "test-app"
	secret := "test-secret-key-32-bytes-long!!"

	// CORS 需要完整 URL，CSRF 只需要主机名
	corsOrigins := []string{
		"https://example.com",
		"https://api.example.org:8080",
	}
	csrfTrustedOrigins := []string{
		"example.com",
		"api.example.org:8080",
	}

	router := gin.New()
	// 先 CORS，再 CSRF
	router.Use(CORS(corsOrigins))
	router.Use(CSRF(appID, secret, csrfTrustedOrigins))
	router.Use(CSRFToken(appID, ""))

	router.GET("/api/token", func(c *gin.Context) {
		token := csrf.Token(c.Request)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	router.POST("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	router.OPTIONS("/api/data", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}

	// Step 1: 获取 CSRF token
	tokenResp, err := client.Get(ts.URL + "/api/token")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, tokenResp.StatusCode)
	defer tokenResp.Body.Close()

	var tokenResponse map[string]string
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	assert.NoError(t, err)
	csrfToken := tokenResponse["token"]

	// Step 2: 发送跨域 POST 请求
	postReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/data", strings.NewReader(`{"key":"value"}`))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Origin", "https://example.com")
	postReq.Header.Set("X-CSRF-Token", csrfToken)

	postResp, err := client.Do(postReq)
	assert.NoError(t, err)
	defer postResp.Body.Close()

	// 应该成功：CORS 允许 https://example.com，CSRF 信任 example.com
	assert.Equal(t, http.StatusOK, postResp.StatusCode)

	// 验证 CORS 响应头
	assert.Equal(t, "https://example.com", postResp.Header.Get("Access-Control-Allow-Origin"))
}

// TestCSRF_AllHTTPMethods 测试所有 HTTP 方法的 CSRF 保护
func TestCSRF_AllHTTPMethods(t *testing.T) {
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

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": c.Request.Method})
	}

	// 安全方法（不需要 CSRF token）
	router.GET("/test", handler)
	router.HEAD("/test", handler)
	router.OPTIONS("/test", handler)

	// 不安全方法（需要 CSRF token）
	router.POST("/test", handler)
	router.PUT("/test", handler)
	router.PATCH("/test", handler)
	router.DELETE("/test", handler)

	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}

	// 获取 CSRF token
	tokenResp, err := client.Get(ts.URL + "/token")
	assert.NoError(t, err)
	defer tokenResp.Body.Close()

	var tokenResponse map[string]string
	_ = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	csrfToken := tokenResponse["token"]

	testCases := []struct {
		method          string
		needToken       bool
		expectWithToken int
		expectNoToken   int
	}{
		// 安全方法：不需要 token
		{"GET", false, http.StatusOK, http.StatusOK},
		{"HEAD", false, http.StatusOK, http.StatusOK},
		{"OPTIONS", false, http.StatusOK, http.StatusOK},
		// 不安全方法：需要 token
		{"POST", true, http.StatusOK, http.StatusForbidden},
		{"PUT", true, http.StatusOK, http.StatusForbidden},
		{"PATCH", true, http.StatusOK, http.StatusForbidden},
		{"DELETE", true, http.StatusOK, http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.method+"_with_token", func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, ts.URL+"/test", nil)
			req.Header.Set("Origin", "https://example.com")
			req.Header.Set("X-CSRF-Token", csrfToken)

			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectWithToken, resp.StatusCode,
				"%s with token should return %d", tc.method, tc.expectWithToken)
		})

		if tc.needToken {
			t.Run(tc.method+"_without_token", func(t *testing.T) {
				req, _ := http.NewRequest(tc.method, ts.URL+"/test", nil)
				req.Header.Set("Origin", "https://example.com")
				// 不设置 CSRF token

				resp, err := client.Do(req)
				assert.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tc.expectNoToken, resp.StatusCode,
					"%s without token should return %d", tc.method, tc.expectNoToken)
			})
		}
	}
}

// TestCSRF_JSONContentType 测试 JSON 请求体的 CSRF 保护
func TestCSRF_JSONContentType(t *testing.T) {
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
	router.POST("/api/resource", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"received": body})
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}

	// 获取 CSRF token
	tokenResp, err := client.Get(ts.URL + "/token")
	assert.NoError(t, err)
	defer tokenResp.Body.Close()

	var tokenResponse map[string]string
	_ = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	csrfToken := tokenResponse["token"]

	// 发送 JSON 请求
	jsonBody := `{"name":"test","value":123}`
	postReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/resource", strings.NewReader(jsonBody))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Accept", "application/json")
	postReq.Header.Set("Origin", "https://example.com")
	postReq.Header.Set("X-CSRF-Token", csrfToken)
	postReq.Header.Set("X-Requested-With", "fetch") // 模拟 AJAX 请求

	postResp, err := client.Do(postReq)
	assert.NoError(t, err)
	defer postResp.Body.Close()

	assert.Equal(t, http.StatusOK, postResp.StatusCode)

	// 验证响应
	var respBody map[string]interface{}
	err = json.NewDecoder(postResp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.NotNil(t, respBody["received"])
}
