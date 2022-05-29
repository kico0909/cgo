package request

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Req struct {
	header   map[string]string `json:"header"`
	respones *http.Response    `json:"respones"`
}

func NewRequest() *Req {
	rs := new(Req)
	rs.SetHeader("Content-type", "application/json")
	return rs
}

func (s *Req) SetHeader(key, value string) {
	if s.header == nil {
		s.header = make(map[string]string)
	}
	s.header[key] = value
}
func (s *Req) SetHeaders(headers map[string]string) {
	if s.header == nil {
		s.header = make(map[string]string)
	}
	for k, v := range headers {
		s.header[k] = v
	}
}
func (s *Req) initHeaders(r *http.Request) {
	for k, v := range s.header {
		r.Header.Set(k, v)
	}
}

func (s *Req) Post(url string, body interface{}, params map[string]string, headers map[string]string) (*http.Response, error) {
	//add post body
	var bodyJson []byte
	var req *http.Request
	if body != nil {
		var err error
		bodyJson, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return nil, err
	}
	if headers != nil {
		s.SetHeaders(headers)
		s.initHeaders(req)
	}

	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//http client
	client := &http.Client{}
	rs, err := client.Do(req)
	if err != nil {
		return rs, err
	}
	s.respones = rs
	return rs, err
}

func (s *Req) Get(url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		s.SetHeaders(headers)
		s.initHeaders(req)
	}
	//http client
	client := &http.Client{}
	rs, err := client.Do(req)
	if err != nil {
		return rs, err
	}
	s.respones = rs
	return rs, err
}

func (s *Req) GetBody() ([]byte, error) {
	defer s.respones.Body.Close()
	return ioutil.ReadAll(s.respones.Body)

}
