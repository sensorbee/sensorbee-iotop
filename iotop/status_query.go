package iotop

import (
	"fmt"

	"gopkg.in/sensorbee/sensorbee.v0/client"
)

// StatusRequester is an interface for streaming node status.
type StatusRequester interface {
	PostQuery(string) (*client.Response, error)
}

type nodeStatusRequester struct {
	req  *client.Requester
	path string
}

func newNodeStatusRequester(addr, ver, tpl string) (StatusRequester, error) {
	req, err := client.NewRequester(addr, ver)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot create a new requester for node monitoring, %v", err)
	}
	path := "/topologies/" + tpl + "/queries"
	return &nodeStatusRequester{
		req:  req,
		path: path,
	}, nil
}

func (n *nodeStatusRequester) PostQuery(bql string) (*client.Response, error) {
	return n.req.Do(client.Post, n.path, map[string]interface{}{
		"queries": bql,
	})
}

func setupStatusQuery(req StatusRequester, interval float64) error {
	createNodeStatusSourceBQL := fmt.Sprintf(
		`CREATE SOURCE iotop_ns TYPE node_statuses WITH interval = %f;`, interval)
	res, err := req.PostQuery(createNodeStatusSourceBQL)
	if err != nil {
		return fmt.Errorf("request failed to create 'node_statuses' source, %v", err)
	}
	defer res.Close()
	if err := checkResponseError(res); err != nil {
		return err
	}
	return nil
}

func selectNodeStatus(req StatusRequester) (res *client.Response, err error) {
	selectNodeStatusBQL := `SELECT RSTREAM *, ts() FROM iotop_ns [RANGE 1 TUPLES];`
	res, err = req.PostQuery(selectNodeStatusBQL)
	if err != nil {
		return nil, fmt.Errorf("request failed to stream 'node_statuses', %v", err)
	}
	defer func() {
		if err != nil {
			res.Close()
		}
	}()
	if err = checkResponseError(res); err != nil {
		return nil, err
	}
	if !res.IsStream() {
		err = fmt.Errorf("failed to stream 'SELECT' query")
		return nil, err
	}
	return
}

func tearDownStatusQuery(req StatusRequester) error {
	_, err := req.PostQuery(`DROP SOURCE iotop_ns;`)
	return err
}

func checkResponseError(res *client.Response) error {
	if res.IsError() {
		errRes, err := res.Error()
		if err != nil {
			return fmt.Errorf("cannot get valid response, %v", err)
		}
		return fmt.Errorf("request failed, %v: %v, %v", errRes.Code,
			errRes.Message, errRes.Meta)
	}
	return nil
}
