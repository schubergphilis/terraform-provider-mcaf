package mcaf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/google/go-querystring/query"
)

var (
	cache map[string]*o365Response
	mutex sync.Mutex
)

func init() {
	cache = make(map[string]*o365Response)
}

type O365Client struct {
	aclGUID    string
	client     *http.Client
	endpoint   *url.URL
	secretCode string
}

type o365Request struct {
	Alias   string `url:"-" json:"proxy_address,omitempty"`
	GroupID string `url:"unifiedgroupid" json:"unified_group_id"`
}

type Status string

const (
	Error      Status = "error"
	InProgress Status = "inprogress"
	Queued     Status = "queued"
	Success    Status = "success"
)

type o365Group struct {
	Aliases []string `json:"proxy_addresses"`
	ID      string   `json:"id"`
}

type o365Response struct {
	Group   o365Group `json:"unified_group"`
	Request struct {
		ID string `json:"id"`
	} `json:"request"`
	Status Status `json:"status"`
}

type o365Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *o365Error) Error() string {
	return fmt.Sprintf("unexpected response: %s (%s)", e.Message, e.Code)
}

func (o *O365Client) Create(groupID string, alias string) (*o365Response, error) {
	req := &o365Request{
		Alias:   alias,
		GroupID: groupID,
	}
	return o.do("POST", req)
}

func (o *O365Client) Read(groupID string) (*o365Response, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if resp, ok := cache[groupID]; ok {
		return resp, nil
	}

	req := &o365Request{
		GroupID: groupID,
	}

	resp, err := o.do("GET", req)
	if err != nil {
		return nil, err
	}
	cache[groupID] = resp

	return resp, nil
}

func (o *O365Client) Delete(groupID string, alias string) (*o365Response, error) {
	req := &o365Request{
		Alias:   alias,
		GroupID: groupID,
	}
	return o.do("DELETE", req)
}

func (o *O365Client) do(method string, opt interface{}) (*o365Response, error) {
	var body io.Reader
	endpoint := *o.endpoint

	switch {
	case method == "POST" || method == "DELETE":
		content, err := json.Marshal(opt)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(content)
	case opt != nil:
		q, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		endpoint.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	resp := new(o365Response)
	if err := o.call(req, resp); err != nil {
		return nil, err
	}

	// Return if there is no request ID to follow up on.
	if resp.Request.ID == "" {
		return resp, nil
	}

	u := fmt.Sprintf("%s/%s", o.endpoint.String(), resp.Request.ID)
	req, err = http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	for resp.Status != Success && resp.Status != Error {
		if err = o.call(req, resp); err != nil {
			return nil, err
		}
		time.Sleep(1 * time.Second)
	}

	return resp, nil
}

func (o *O365Client) call(req *http.Request, v interface{}) error {
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-acl-guid", o.aclGUID)
	req.Header.Set("x-functions-key", o.secretCode)

	request, err := httputil.DumpRequest(req, true)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] request: %s", string(request))

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	response, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] response: %s", string(response))

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		var e struct {
			Error *o365Error `json:"error"`
		}
		e.Error = &o365Error{
			Code: resp.Status,
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			e.Error.Message = err.Error()
			return e.Error
		}
		e.Error.Message = string(body)

		// Try to unmarshal the body, but ignore any errors as we
		// already set the complete body as default error message.
		_ = json.Unmarshal(body, &e)

		return e.Error
	}

	return json.NewDecoder(resp.Body).Decode(v)
}
