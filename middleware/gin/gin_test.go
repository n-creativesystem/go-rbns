package gin_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	rbns "github.com/n-creativesystem/go-rbns"
	rbnsGin "github.com/n-creativesystem/go-rbns/middleware/gin"
	"github.com/n-creativesystem/go-rbns/tests"
	"github.com/stretchr/testify/assert"
)

func getUser(c *gin.Context) (userKey, organizationName string, err error) {
	userKey = c.GetHeader("X-User")
	organizationName = "default"
	err = nil
	return
}

func getUserNoExistsOrganization(c *gin.Context) (userKey, organizationName string, err error) {
	userKey = c.GetHeader("X-User")
	organizationName = "default2"
	err = nil
	return
}

func middlewarePermission(permissions ...string) gin.HandlerFunc {
	return rbnsGin.PermissionCheck(getUser, permissions...)
}

func middlewarePermissionNoExistsOrganization(permissions ...string) gin.HandlerFunc {
	return rbnsGin.PermissionCheck(getUserNoExistsOrganization, permissions...)
}

func request(method, url, userKey string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	req.Header.Set("X-User", userKey)
	return req
}

func TestGinWAF(t *testing.T) {
	router := gin.New()
	router.Use(rbnsGin.ClientWithOptions(rbns.WithHost(fmt.Sprintf("%s:%d", "api-rbac-dev", 8888)), rbns.WithApiKey("5d78ced0-c6a0-471d-90f3-a0ec7653172e")))
	users := router.Group("/api")
	{
		users.POST("/users", middlewarePermission("create:test"), func(c *gin.Context) {
			userKey := c.GetHeader("X-User")
			c.JSON(http.StatusOK, gin.H{"user": userKey})
		})
		users.GET("/users/:id", middlewarePermission("read:test"), func(c *gin.Context) {
			userKey := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"user": userKey})
		})
		users.DELETE("/users/:id", middlewarePermissionNoExistsOrganization("delete:test"), func(c *gin.Context) {
			userKey := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"user": userKey})
		})
	}
	noUsers := router.Group("/api")
	{
		noUsers.POST("/no-users", middlewarePermissionNoExistsOrganization("create:test"), func(c *gin.Context) {
			userKey := c.GetHeader("X-User")
			c.JSON(http.StatusOK, gin.H{"user": userKey})
		})
		noUsers.GET("/no-users/:id", middlewarePermissionNoExistsOrganization("read:test"), func(c *gin.Context) {
			userKey := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"user": userKey})
		})
		noUsers.DELETE("/no-users/:id", middlewarePermissionNoExistsOrganization("delete:test"), func(c *gin.Context) {
			userKey := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"user": userKey})
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
										assert.Equal(t, `{"user":"user1"}`, w.Body.String())
									},
								},
								{
									Name: "read ok",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/users/user2", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusOK, w.Result().StatusCode)
										assert.Equal(t, `{"user":"user2"}`, w.Body.String())
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/users/user2", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
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
										assert.Empty(t, w.Body.String())
									},
								},
								{
									Name: "read ok",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/users/user1", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusOK, w.Result().StatusCode)
										assert.Equal(t, `{"user":"user1"}`, w.Body.String())
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/users/user1", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
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
										assert.Empty(t, w.Body.String())
									},
								},
								{
									Name: "read ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/users/user1", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/users/user1", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
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
										assert.Empty(t, w.Body.String())
									},
								},
								{
									Name: "read ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/no-users/user2", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/no-users/user2", "user1")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
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
										assert.Empty(t, w.Body.String())
									},
								},
								{
									Name: "read ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/no-users/user1", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/no-users/user1", "user2")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
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
										assert.Empty(t, w.Body.String())
									},
								},
								{
									Name: "read ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodGet, "/api/no-users/user1", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
									},
								},
								{
									Name: "delete ng",
									Fn: func(t *testing.T) {
										req = request(http.MethodDelete, "/api/no-users/user1", "user3")
										w = httptest.NewRecorder()
										router.ServeHTTP(w, req)
										assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
										assert.Empty(t, w.Body.String())
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
