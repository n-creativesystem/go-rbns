package middleware

import (
	"errors"
	"net/http"

	rbns "github.com/n-creativesystem/sdk-go-rbns"
)

var (
	ErrForbidden = errors.New(http.StatusText(http.StatusForbidden))
)

func PermissionCheck(client *rbns.Client, userKey, organizationName string, permissionNames ...string) error {
	r, err := client.Check(userKey, organizationName, permissionNames...)
	if err != nil {
		return err
	}
	if !r {
		return ErrForbidden
	}
	return nil
}
