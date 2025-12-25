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

	router := gin.New()
	router.Use(CSRF(appID, secret))
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

	router := gin.New()
	router.Use(CSRF(appID, secret))
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

	// Create a router with both CSRF protection and token setting
	router := gin.New()
	router.Use(CSRF(appID, secret))
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

	// CSRFToken requires CSRF middleware to be applied first
	router := gin.New()
	router.Use(CSRF(appID, secret))
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

	router := gin.New()
	router.Use(CSRF(appID, secret))
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
