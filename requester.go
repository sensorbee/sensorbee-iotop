package main

import (
	"fmt"

	"gopkg.in/sensorbee/sensorbee.v0/client"
	"gopkg.in/urfave/cli.v1"
)

func newRequester(c *cli.Context) (*client.Requester, error) {
	r, err := client.NewRequester(c.String("uri"), c.String("api-version"))
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP requester, %v", err)
	}
	return r, nil
}

func do(c *cli.Context, method client.Method, path string, body interface{},
	baseErrMsg string) (*client.Response, error) {
	req, err := newRequester(c)
	if err != nil {
		return nil, fmt.Errorf("%v, %v", baseErrMsg, err)
	}
	res, err := req.Do(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("%v, %v", baseErrMsg, err)
	}
	if res.IsError() {
		errRes, err := res.Error()
		if err != nil {
			return nil, fmt.Errorf("%v and failed to parse error information, %v",
				baseErrMsg, err)
		}
		return nil, fmt.Errorf("%v: %v, %v: %v", baseErrMsg, errRes.Code,
			errRes.RequestID, errRes.Message)
	}
	return res, nil
}
