package iotop

import (
	"fmt"

	"gopkg.in/sensorbee/sensorbee.v0/client"
)

type nodeStatuser struct {
	req  *client.Requester
	res  *client.Response
	path string
	stop chan struct{}
}

func newNodeStatuser(addr, apiVer, tpl string) (*nodeStatuser, error) {
	req, err := client.NewRequester(addr, apiVer)
	if err != nil {
		return nil, fmt.Errorf("cannot create new requester, %v", err)
	}
	path := "/topologies/" + tpl + "/queries"
	return &nodeStatuser{
		req:  req,
		path: path,
		stop: make(chan struct{}, 1),
	}, nil
}

func (n *nodeStatuser) selectNodeStatus(interval float64) (
	<-chan interface{}, error) {
	createNodeStatusSourceBQL := fmt.Sprintf(
		`CREATE SOURCE iotop_ns TYPE node_statuses WITH interval = %f;`, interval)
	res, err := n.postQuery(createNodeStatusSourceBQL)
	if err != nil {
		return nil, fmt.Errorf("request failed to create 'node_statuses' source, %v", err)
	}
	defer res.Close()
	if err := checkResponseError(res); err != nil {
		return nil, err
	}

	selectNodeStatusBQL := `SELECT RSTREAM *, ts() FROM iotop_ns [RANGE 1 TUPLES];`
	sres, err := n.postQuery(selectNodeStatusBQL)
	if err != nil {
		return nil, fmt.Errorf("request failed to stream 'node_statuses', %v", err)
	}

	var serr error
	defer func() {
		if serr != nil {
			sres.Close()
		}
	}()
	if serr = checkResponseError(sres); serr != nil {
		return nil, serr
	}
	if !sres.IsStream() {
		serr = fmt.Errorf("failed to stream 'SELECT' query")
		return nil, serr
	}
	ch, err := sres.ReadStreamJSON()
	if err != nil {
		serr = fmt.Errorf("cannot read stream channel, %v", err)
		return nil, serr
	}
	n.res = sres
	return ch, nil
}

func (n *nodeStatuser) terminate() {
	// skip catch error
	defer n.res.Close()
	dropNodeStatusSouceBQL := `DROP SOURCE iotop_ns;`
	n.postQuery(dropNodeStatusSouceBQL)
}

func (n *nodeStatuser) postQuery(bql string) (*client.Response, error) {
	return n.req.Do(client.Post, n.path, map[string]interface{}{
		"queries": bql,
	})
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
