package proxy

import "fmt"
import "net"
import "net/http"
import "io/ioutil"

type UpStream struct {
	Name string;
	handle *http.Client;
}

func NewSocket(socket string) UpStream {
	stream := UpStream{ Name: socket };
	stream.handle = &http.Client{
		Transport: &http.Transport{
			Dial: func(proto, addr string) (net.Conn,error) {
				conn, err := net.Dial("unix", socket);
				return conn, err;
			},
		},
	}
	return stream;
}

func (r UpStream) Pass() (func(res http.ResponseWriter, req *http.Request)) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(res, "400 Bad request ; only GET allowed.", 400)
			return;
		}
		param := "";
		if len(req.URL.RawQuery) > 0 {
			param = "?" + req.URL.RawQuery;
		}
		body, _ := r.Get("http://docker" + req.URL.Path + param, res);
		fmt.Fprintf(res, "%s", body);
	}
}

func (r UpStream) Get(url string, res http.ResponseWriter) ([]byte,error) {
	req, err := r.handle.Get(url);
	if err != nil {
		return nil, err
	}
	defer req.Body.Close();
	contentType := req.Header.Get("Content-type");
	if contentType != "" {
		res.Header().Set("Content-type", contentType);
	}
	return ioutil.ReadAll(req.Body)
}
