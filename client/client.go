package client

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/dchest/uniuri"
	"github.com/lavab/api/models"
	"github.com/lavab/api/routes"
	"github.com/lavab/sockjs-go-client"
)

type Client struct {
	sync.RWMutex

	Address  string
	SockJS   *sockjs.Client
	Headers  map[string]string
	Incoming map[string]chan *Response
	Timeout  time.Duration
}

func New(address string, timeout time.Duration) (*Client, error) {
	sjs, err := sockjs.NewClient(address + "/ws")
	if err != nil {
		return nil, err
	}

	if timeout == 0 {
		timeout = time.Second * 30
	}

	client := &Client{
		Address:  address,
		SockJS:   sjs,
		Headers:  map[string]string{},
		Incoming: map[string]chan *Response{},
		Timeout:  timeout,
	}

	go client.Loop()

	return client, nil
}

func (c *Client) Loop() {
	for {

		x := []string{}
		if err := c.SockJS.ReadMessage(&x); err != nil {
			log.Print(err)
			break
		}

		var resp *Response
		if err := json.Unmarshal([]byte(x[0]), &resp); err != nil {
			log.Print(err)
			continue
		}

		if resp.Type == "response" {
			c.RLock()
			d, ok := c.Incoming[resp.ID]
			c.RUnlock()

			if ok {
				d <- resp
			}
		} else {
			// Run event handlers
		}
	}
}

func (c *Client) Receive(id string) (*Response, error) {
	c.Lock()
	c.Incoming[id] = make(chan *Response)
	c.Unlock()

	select {
	case <-time.After(c.Timeout):
		return nil, errors.New("Request timeout")
	case data := <-c.Incoming[id]:
		return data, nil
	}

	return nil, errors.New("This shouldn't happen!")
}

func (c *Client) Request(method, path string, headers map[string]string, body interface{}) ([]string, string, error) {
	if headers == nil {
		headers = map[string]string{}
	}

	if c.Headers != nil {
		for k, v := range c.Headers {
			headers[k] = v
		}
	}

	req := &Request{
		ID:      uniuri.New(),
		Type:    "request",
		Path:    path,
		Method:  method,
		Headers: headers,
	}

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, "", err
		}

		req.Body = string(data)
	}

	d, err := Encode(req)
	if err != nil {
		return nil, "", err
	}

	return d, req.ID, nil
}

func (c *Client) CreateToken(req *routes.TokensCreateRequest) (*models.Token, error) {
	data, id, err := c.Request("POST", "/tokens", map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}, req)
	if err != nil {
		return nil, err
	}

	if err := c.SockJS.WriteMessage(data); err != nil {
		return nil, err
	}

	rcv, err := c.Receive(id)
	if err != nil {
		return nil, err
	}

	var resp *routes.TokensCreateResponse
	if err := json.Unmarshal([]byte(rcv.Body), &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, errors.New(resp.Message)
	}

	return resp.Token, nil
}

func (c *Client) CreateEmail(req *routes.EmailsCreateRequest) ([]string, error) {
	data, id, err := c.Request("POST", "/emails", map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}, req)
	if err != nil {
		return nil, err
	}

	if err := c.SockJS.WriteMessage(data); err != nil {
		return nil, err
	}

	rcv, err := c.Receive(id)
	if err != nil {
		return nil, err
	}

	var resp *routes.EmailsCreateResponse
	if err := json.Unmarshal([]byte(rcv.Body), &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, errors.New(resp.Message)
	}

	return resp.Created, nil
}

func (c *Client) GetKey(id string) (*models.Key, error) {
	data, id, err := c.Request("GET", "/keys/"+id, nil, nil)
	if err != nil {
		return nil, err
	}

	if err := c.SockJS.WriteMessage(data); err != nil {
		return nil, err
	}

	rcv, err := c.Receive(id)
	if err != nil {
		return nil, err
	}

	var resp *routes.KeysGetResponse
	if err := json.Unmarshal([]byte(rcv.Body), &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, errors.New(resp.Message)
	}

	return resp.Key, nil
}
