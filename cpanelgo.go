package cpanelgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	megabyte          = 1 * 1024 * 1024
	ResponseSizeLimit = (5 * megabyte) + 1337
)

type UAPIResult struct {
	ApiVersion int             `json:"apiversion"`
	Func       string          `json:"func"`
	Module     string          `json:"module"`
	Result     json.RawMessage `json:"result"`
}

type API2Result struct {
	Result json.RawMessage `json:"cpanelresult"`
}

type BaseUAPIResponse struct {
	StatusCode int      `json:"status"`
	Errors     []string `json:"errors"`
	Messages   []string `json:"messages"`
}

func (r BaseUAPIResponse) Error() error {
	if r.StatusCode == 1 {
		return nil
	}
	if len(r.Errors) == 0 {
		return errors.New("Unknown")
	}
	return errors.New(strings.Join(r.Errors, "\n"))
}

func (r BaseUAPIResponse) Message() error {
	if r.Messages == nil || len(r.Messages) == 0 {
		return nil
	}
	return errors.New(strings.Join(r.Messages, "\n"))
}

type BaseAPI2Response struct {
	Event struct {
		Result int    `json:"result"`
		Reason string `json:"reason"`
	} `json:"event"`
}

func (r BaseAPI2Response) Error() error {
	if r.Event.Result == 1 {
		return nil
	}
	if len(r.Event.Reason) == 0 {
		return errors.New("Unknown")
	}
	return errors.New(r.Event.Reason)
}

type BaseAPI1Response struct {
	// other stuff here "apiversion":"1","type":"event","module":"Serverinfo","func":"servicestatus","source":"module"
	Data struct {
		Result string `json:"result"`
	} `json:"data"`
	ErrorString string `json:"error"`
	// "event":{"result":1,"reason":"blah blah"}}
	Event struct {
		Result int    `json:"result"`
		Reason string `json:"reason"`
	} `json:"event"`
}

func (r BaseAPI1Response) Error() error {
	if r.ErrorString != "" {
		return errors.New(r.ErrorString)
	}
	if r.Event.Result != 1 {
		// if the result != 1 the reason usually present in error ^ so kinda redundant to check, but check just in case
		if len(r.Event.Reason) == 0 {
			return errors.New("Unknown")
		}
		return errors.New(r.Event.Reason)
	}
	return nil
}

type Args map[string]interface{}

func (a Args) Values(apiVersion string) (url.Values) {
	vals := url.Values{}
	for k, v := range a {
		if apiVersion == "1" {
			kv := strings.SplitN(k, "=", 2)
			if len(kv) == 1 {
				vals.Add(kv[0], "")
			} else if len(kv) == 2 {
				vals.Add(kv[0], kv[1])
			}
		} else {
			vals.Add(k, fmt.Sprintf("%v", v))
		}
	}
	return vals
}

type ApiGateway interface {
	UAPI(module, function string, arguments Args, out interface{}) error
	API2(module, function string, arguments Args, out interface{}) error
	API1(module, function string, arguments []string, out interface{}) error
	Close() error
}

type Api struct {
	Gateway ApiGateway
}

func NewApi(gw ApiGateway) Api {
	return Api{
		Gateway: gw,
	}
}

func (a Api) Close() error {
	if a.Gateway != nil {
		return a.Gateway.Close()
	} else {
		return nil
	}
}