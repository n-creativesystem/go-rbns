package fwncs_test

import (
	"fmt"
	"n-creativesystem/go-rbns-sdk"
	rbnsFwncs "n-creativesystem/go-rbns-sdk/middleware/fwncs"
	"n-creativesystem/go-rbns-sdk/tests"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/n-creativesystem/go-fwncs"
	"github.com/stretchr/testify/assert"
)

func getUser(c fwncs.Context) (userKey, organizationName string, err error) {
	userKey = c.Header().Get("X-User")
	organizationName = "default"
	err = nil
	return
}

func getUserNoExistsOrganization(c fwncs.Context) (userKey, organizationName string, err error) {
	userKey = c.Header().Get("X-User")
	organizationName = "default2"
	err = nil
	return
}

func middlewarePermission(permissions ...string) fwncs.HandlerFunc {
	return rbnsFwncs.PermissionCheck(getUser, permissions...)
}

func middlewarePermissionNoExistsOrganization(permissions ...string) fwncs.HandlerFunc {
	return rbnsFwncs.PermissionCheck(getUserNoExistsOrganization, permissions...)
}

func request(method, url, userKey string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	req.Header.Set("X-User", userKey)
	return req
}

func TestFwncsWAF(t *testing.T) {
	router := fwncs.New()
	router.Use(rbnsFwncs.ClientWithOptions(rbns.WithHost(fmt.Sprintf("%s:%d", "api-rbac-dev", 8888)), rbns.WithApiKey("5d78ced0-c6a0-471d-90f3-a0ec7653172e")))
	users := router.Group("/api")
	{
		users.POST("/users", middlewarePermission("create:test"), func(c fwncs.Context) {
			userKey := c.Header().Get("X-User")
			c.JSON(http.StatusOK, map[string]interface{}{"user": userKey})
		})
		users.GET("/users/:id", middlewarePermission("read:test"), func(c fwncs.Context) {
			userKey := c.Param("id")
			c.JSON(http.StatusOK, map[string]interface{}{"user": userKey})
		})
		users.DELETE("/users/:id", middlewarePermissionNoExistsOrganization("delete:test"), func(c fwncs.Context) {
			userKey := c.Param("id")
			c.JSON(http.StatusOK, map[string]interface{}{"user": userKey})
		})
	}
	noUsers := router.Group("/api")
	{
		noUsers.POST("/no-users", middlewarePermissionNoExistsOrganization("create:test"), func(c fwncs.Context) {
			userKey := c.Header().Get("X-User")
			c.JSON(http.StatusOK, map[string]interface{}{"user": userKey})
		})
		noUsers.GET("/no-users/:id", middlewarePermissionNoExistsOrganization("read:test"), func(c fwncs.Context) {
			userKey := c.Param("id")
			c.JSON(http.StatusOK, map[string]interface{}{"user": userKey})
		})
		noUsers.DELETE("/no-users/:id", middlewarePermissionNoExistsOrganization("delete:test"), func(c fwncs.Context) {
			userKey := c.Param("id")
			c.JSON(http.StatusOK, map[string]interface{}{"user": userKey})
		})
	}
	var req *http.Request
	var w *httptest.ResponseRecorder

	cases := tests.Cases{
		{
			Name: "organization ok",
			Fn: func(t *testing.T) {
				childCases := tests.Cases{
					{
						Name: "user1",
						Fn: func(t *testing.T) {
							childCases := tests.Cases{
								{
									Name: "create ok",
									Fn: func(t *testing.T) {
										req = request(http.MethodPost, "/api/users", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusOK, w.Result().StatusCode)
										assert.Equal(t, `{"user":"user1"}`, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "read ok",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/users/user2", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusOK, w.Result().StatusCode)
										assert.Equal(t, `{"user":"user2"}`, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/users/user2", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
							}
							childCases.Run(t)
						},
					},
					{
						Name: "user2",
						Fn: func(t *testing.T) {
							childCases := tests.Cases{
								{
									Name: "create ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodPost, "/api/users", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "read ok",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/users/user1", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusOK, w.Result().StatusCode)
										assert.Equal(t, `{"user":"user1"}`, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/users/user1", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
							}
							childCases.Run(t)
						},
					},
					{
						Name: "user3 no exists",
						Fn: func(t *testing.T) {
							childCases := tests.Cases{
								{
									Name: "create ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodPost, "/api/users", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "read ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/users/user1", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/users/user1", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
							}
							childCases.Run(t)
						},
					},
				}
				childCases.Run(t)
			},
		},
		{
			Name: "organization no exists",
			Fn: func(t *testing.T) {
				childCases := tests.Cases{
					{
						Name: "user1",
						Fn: func(t *testing.T) {
							childCases := tests.Cases{
								{
									Name: "create ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodPost, "/api/no-users", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "read ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/no-users/user2", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/no-users/user2", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
							}
							childCases.Run(t)
						},
					},
					{
						Name: "user2",
						Fn: func(t *testing.T) {
							childCases := tests.Cases{
								{
									Name: "create ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodPost, "/api/no-users", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "read ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/no-users/user1", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/no-users/user1", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
							}
							childCases.Run(t)
						},
					},
					{
						Name: "user3 no exists",
						Fn: func(t *testing.T) {
							childCases := tests.Cases{
								{
									Name: "create ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodPost, "/api/no-users", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "read ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/no-users/user1", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/no-users/user1", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.NotEmpty(t, strings.ReplaceAll(w.Body.String(), "\n", ""))
									},
								},
							}
							childCases.Run(t)
						},
					},
				}
				childCases.Run(t)
			},
		},
	}
	cases.Run(t)
}
