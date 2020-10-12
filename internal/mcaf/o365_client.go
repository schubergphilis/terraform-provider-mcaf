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
	Aliases []string `json:"MailNicknames,omitempty"`
	GroupID string   `json:"UnifiedGroupID"`
}

type o365StatusURL struct {
	StatusURL string `json:"StatusURL"`
}

type Status string

const (
	Done       Status = "Done"
	Error      Status = "Error"
	InProgress Status = "InProgress"
	Queued     Status = "Queued"
)

type o365Response struct {
	Aliases []string `json:"MailNicknames"`
	GroupID string   `json:"UnifiedGroupID"`
	Status  Status   `json:"Status"`
}

type o365Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *o365Error) Error() string {
	return fmt.Sprintf("unexpected response: %s (%s)", e.Message, e.Code)
}

func (o *O365Client) Create(groupID string, aliases []string) (*o365Response, error) {
	req := &o365Request{
		Aliases: aliases,
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

func (o *O365Client) Delete(groupID string, aliases []string) (*o365Response, error) {
	req := &o365Request{
		Aliases: aliases,
		GroupID: groupID,
	}
	return o.do("DELETE", req)
}

func (o *O365Client) do(method string, reqBody *o365Request) (*o365Response, error) {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(bodyBytes)

	u := fmt.Sprintf("%s?code=%s", o.endpoint.String(), o.secretCode)
	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, err
	}

	status := new(o365StatusURL)
	if err := o.call(req, status); err != nil {
		return nil, err
	}

	u = fmt.Sprintf("%s&code=%s", status.StatusURL, o.secretCode)
	req, err = http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp := new(o365Response)
	for resp.Status != Done && resp.Status != Error {
		if err = o.call(req, resp); err != nil {
			return nil, err
		}
		time.Sleep(5 * time.Second)
	}

	return resp, nil
}

func (o *O365Client) call(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Aclguid", o.aclGUID)

	request, err := httputil.DumpRequest(req, true)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] %s", string(request))

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	response, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] %s", string(response))

	if resp.StatusCode != http.StatusOK {
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
