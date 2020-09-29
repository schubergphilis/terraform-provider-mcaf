package mcaf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
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
	Alias   string `json:"proxy_address,omitempty"`
	GroupID string `json:"unified_group_id"`
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
	Id      string   `json:"id"`
}

type o365Response struct {
	Group   o365Group `json:"unified_group,omitempty"`
	Request struct {
		Id string `json:"id"`
	} `json:"Request,omitempty"`
	Status Status `json:"Status,omitempty"`
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
	return o.do("POST", o.endpoint.String(), req)
}

func (o *O365Client) Read(groupID string) (*o365Response, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if resp, ok := cache[groupID]; ok {
		return resp, nil
	}

	u := fmt.Sprintf("%s?unifiedgroupid=%s", o.endpoint.String(), groupID)
	resp, err := o.do("GET", u, nil)
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
	return o.do("DELETE", o.endpoint.String(), req)
}

func (o *O365Client) do(method string, endpoint string, reqBody *o365Request) (*o365Response, error) {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest(method, endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	resp := new(o365Response)
	if err := o.call(req, resp); err != nil {
		return nil, err
	}

	// Return if there is no request ID to follow up on.
	if resp.Request.Id == "" {
		return resp, nil
	}

	u := fmt.Sprintf("%s/%s", endpoint, resp.Request.Id)
	req, err = http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	for resp.Status != Success && resp.Status != Error {
		if err = o.call(req, resp); err != nil {
			return nil, err
		}
		time.Sleep(5 * time.Second)
	}

	return resp, nil
}

func (o *O365Client) call(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ACL-GUID", o.aclGUID)
	req.Header.Set("X-Functions-Key", o.secretCode)

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
