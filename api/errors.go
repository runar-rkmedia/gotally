package api

import (
	"github.com/bufbuild/connect-go"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type CError struct {
	connect.Error
}

func (c *CError) ToConnectError() *connect.Error {
	return &c.Error
}
func (c *CError) AddBadRequestDetail(violations []*errdetails.BadRequest_FieldViolation) *CError {

	details := &errdetails.BadRequest{
		FieldViolations: violations,
	}
	if details, detailErr := connect.NewErrorDetail(details); detailErr == nil {
		c.AddDetail(details)
	}
	return c
}

func createError(c connect.Code, err error) CError {
	cerr := CError{*connect.NewError(c, err)}
	return cerr
}
