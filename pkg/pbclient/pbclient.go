package pbclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/davesavic/clink"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
)

const (
	EndpointAuthAdmin = "/api/admin/auth-with-password"
	EndpointAuthUser  = "/api/collections/users/auth-with-password"
	EndpointRecords   = "/api/collections/%s/records"
	EndpointRecordID  = "/api/collections/%s/records/%s"
)

var ErrInvalidStatusCode = errors.New("invalid status code")

func getEndpoint(endpoint string, params ...any) string {
	return fmt.Sprintf(endpoint, params...)
}

type (
	Client struct {
		BaseURL string
		Token   string
		Client  *clink.Client
	}
	Payload = map[string]string
)

func New(baseURL string) Client {
	c := clink.NewClient()
	c.Headers["Content-Type"] = "application/json"
	return Client{BaseURL: baseURL, Client: c}
}

func (c Client) buildUrl(path string, qms ...IQ) string {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		xlog.Error("failed to parse base url", "baseURL", c.BaseURL)
		return c.BaseURL
	}

	var params = []any{}
	for _, qm := range qms {
		if qm.GetQType() == QTypeParams {
			params = qm.(QParams)
		}
	}
	parsedPath, err := url.Parse(getEndpoint(path, params...))
	if err != nil {
		xlog.Error("failed to parse path", "path", path)
		return c.BaseURL
	}
	u := base.ResolveReference(parsedPath)
	var (
		q = u.Query()
	)
	for _, qm := range qms {
		switch qm := qm.(type) {
		case QExpand:
			q.Set("expand", QmListString(qm))
		case QFields:
			q.Set("fields", QmListString(qm))
		case QSort:
			q.Set("sort", QmListString(qm))
		case QFilters:
			q.Set("filters", string(qm))
		case QPage:
			q.Set("page", strconv.Itoa(qm.Page))
			q.Set("perPage", strconv.Itoa(qm.PerPage))
			if qm.SkipTotal {
				q.Set("skipTotal", "true")
			}
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *Client) Request(method, url string, args ...IQ) (*http.Response, error) {
	url = c.buildUrl(url, args...)
	var (
		body    QData
		headers = QHeaders{}
	)
	for _, arg := range args {
		switch arg := arg.(type) {
		case QData:
			body = arg
		case QHeaders:
			headers = arg
		}
	}
	xlog.Debug("request attr", "url", url, "headers", headers, "payload", body.Data)
	req, err := http.NewRequest(method, url, body.DataBytes)
	for key, val := range headers {
		req.Header.Set(key, val)
	}
	if err != nil {
		xlog.Error("failed to prepare request", "url", url)
		return nil, err
	}
	xlog.Debug("calling", "url", req.URL.String())
	resp, err := c.Client.Do(req)
	if err != nil {
		xlog.Error("error while trying to fetch url", "url", url, "status-code", resp.StatusCode)
		return resp, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, ErrInvalidStatusCode
	}
	return resp, nil
}

func (c *Client) Auth(endpoint, username, password string) (*http.Response, error) {
	payload := RequestAuth{
		Identity: username,
		Password: password,
	}
	return c.Request(http.MethodPost, endpoint, NewQData(payload))
}

func (c *Client) AdminAuth(username, password string) (*http.Response, error) {
	return c.Auth(EndpointAuthAdmin, username, password)
}

func (c *Client) UserAuth(username, password string) (*http.Response, error) {
	return c.Auth(EndpointAuthUser, username, password)
}

func (c *Client) RecordCreate(name string, qs ...IQ) (*http.Response, error) {
	return c.Request(http.MethodPost, EndpointRecords, qs...)
}

func (c *Client) RecordGet(name string, id string, qs ...IQ) (*http.Response, error) {
	return c.Request(http.MethodGet, EndpointRecordID, append(qs, QParams{name, id})...)
}

func (c *Client) RecordList(name string, qs ...IQ) (*http.Response, error) {
	return c.Request(http.MethodGet, EndpointRecords, append(qs, QParams{name})...)
}

func (c *Client) RecordUpdate(name string, id string, qs ...IQ) (*http.Response, error) {
	return c.Request(http.MethodPatch, EndpointRecordID, append(qs, QParams{name, id})...)
}

func (c *Client) RecordDelete(name string, id string) (*http.Response, error) {
	return c.Request(http.MethodDelete, EndpointRecordID, QParams{name, id})
}

func ResponseTo[T any](resp *http.Response) T {
	var t T
	if err := clink.ResponseToJson(resp, &t); err != nil {
		xlog.Error("failed to unmarshal response", "t", t)
	}
	return t
}

func jsonMarshal(d any) io.Reader {
	b, err := json.Marshal(d)
	if err != nil {
		xlog.Error("failed to marshal payload", "payload", d)
	}
	return bytes.NewReader(b)
}
