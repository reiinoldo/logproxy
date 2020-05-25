package handlers

import (
	"fmt"
	"github.com/philips-software/go-hsdp-api/logging"
	"github.com/philips-software/logproxy/queue"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/influxdata/go-syslog/v2/rfc5424"
	"github.com/labstack/echo"
)

var (
	ironIOPayloadRegex = regexp.MustCompile(`severity=(?P<severity>[^\?,]+), task_id: (?P<taskID>[^\?,]+), code_name: (?P<codeName>[^\?,]+), project_id: (?P<projectID>[^\?\s]+) -- (?P<body>.*)`)
)

type IronIOHandler struct {
	pusher queue.Queue
	debug  bool
	token  string
}

func NewIronIOHandler(token string, pusher queue.Queue) (*IronIOHandler, error) {
	if token == "" {
		return nil, fmt.Errorf("Missing TOKEN value")
	}
	handler := &IronIOHandler{}
	handler.token = token
	handler.pusher = pusher

	if os.Getenv("DEBUG") == "true" {
		handler.debug = true
	}
	return handler, nil
}

func ironToRFC5424(now time.Time, ironString string) string {
	msg := &rfc5424.SyslogMessage{}

	msg.SetPriority(14)
	msg.SetVersion(1)
	msg.SetTimestamp(now.Format(logging.TimeFormat))

	match := ironIOPayloadRegex.FindStringSubmatch(ironString)
	if match != nil {
		if len(match) == 6 {
			msg.SetProcID(match[2])
			msg.SetAppname(match[3])
			msg.SetHostname(match[4])
			msg.SetMessage(match[5])
		} else {
			msg.SetMessage("mismatch: " + ironString)
		}
	} else {
		msg.SetMessage("nomatch: " + ironString) // Naive
	}

	out, _ := msg.String()
	return out
}

func (h *IronIOHandler) Handler() echo.HandlerFunc {
	return func(c echo.Context) error {
		t := c.Param("token")
		if h.token != t {
			return c.String(http.StatusUnauthorized, "")
		}
		b, _ := ioutil.ReadAll(c.Request().Body)
		now := time.Now().UTC()
		go func() {
			_ = h.pusher.Push([]byte(ironToRFC5424(now, string(b))))
		}()
		return c.String(http.StatusOK, "")
	}
}