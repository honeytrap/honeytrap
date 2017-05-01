package proxies

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"time"

	config "github.com/honeytrap/honeytrap/config"
	director "github.com/honeytrap/honeytrap/director"
	proxies "github.com/honeytrap/honeytrap/proxies"
	pushers "github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
	"github.com/satori/go.uuid"
)

var log = logging.MustGetLogger("honeytrap:proxy:sip")

// ListenHTTP returns a new listener to handle proxy connections through the
// http protocol.
// TODO: Change amount of params.
func ListenHTTP(address string, d director.Director, p *pushers.Pusher, e pushers.Events, c *config.Config) (net.Listener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	return &HTTPProxyListener{
		proxies.NewProxyListener(l, d, p, e),
	}, nil
}

// HTTPProxyListener defines the struct which handles the underline operation for
// proxying between two http request.
type HTTPProxyListener struct {
	*proxies.ProxyListener
}

// Accept returns a new http connection which handles the proxying operations
// for that request.
func (l *HTTPProxyListener) Accept() (net.Conn, error) {
	conn, err := l.ProxyListener.Accept()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &HTTPProxyConn{conn.(*proxies.ProxyConn)}, err
}

// HTTPProxyConn defines a connection struct for the http connection.
type HTTPProxyConn struct {
	*proxies.ProxyConn
}

// HTTPRequest defines a data struct for the recieved http request.
type HTTPRequest struct {
	Date          time.Time           `json:"timestamp"`
	Host          string              `json:"host"`
	URL           string              `json:"url"`
	RemoteAddr    string              `json:"remote_addr"`
	Method        string              `json:"method"`
	Referer       string              `json:"referer"`
	UserAgent     string              `json:"user_agent"`
	Body          string              `json:"body"`
	ContentLength int64               `json:"content_length"`
	Headers       map[string][]string `json:"headers"`
}

// Proxy initializes and proxies the underline connection operation.
func (p *HTTPProxyConn) Proxy() error {
	sessionID := uuid.NewV4()

	for {
		reader := bufio.NewReader(p.Conn)
		req, err := http.ReadRequest(reader)
		if err != nil {
			return err
		}

		log.Debug("Request for url: %s%s", req.Host, req.URL.String())

		reqBody := &bytes.Buffer{}

		defer func() {
			p.Pusher.Push("http", "request", p.Container.Name(), sessionID.String(), HTTPRequest{
				Date:          time.Now(),
				Host:          req.Host,
				URL:           req.URL.String(),
				RemoteAddr:    p.RemoteHost(),
				Method:        req.Method,
				UserAgent:     req.UserAgent(),
				Referer:       req.Referer(),
				Headers:       req.Header,
				ContentLength: req.ContentLength,
				Body:          reqBody.String(),
			})
		}()

		// dsw := io.MultiWriter(p.server, NewRecordSessionWriter("HTTP-Request-Data", rs))
		dsw := io.MultiWriter(p.Server, reqBody)

		if err = req.Write(dsw); err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		var resp *http.Response
		reader2 := bufio.NewReader(p.Server)
		resp, err = http.ReadResponse(reader2, req)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		// psw := io.MultiWriter(p.Conn, NewRecordSessionWriter("HTTP-Response-Data", rs))

		err = resp.Write(p.Conn)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}

}
