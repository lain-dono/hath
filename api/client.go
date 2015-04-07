package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	http.Client
}

func newClient() *Client {
	tr := &http.Transport{}
	return &Client{
		Client: http.Client{
			//Timeout:   3600 * time.Second, //3600 000
			Timeout:   time.Minute, //3600 000
			Transport: tr,
		},
	}
}

func (client *Client) _Get(url, retryAct string, retryHandler *API) (resp string, err error) {
	r, err := client.Get(url)
	if err != nil {
		log.Error("xx", url)
		return
	}
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	split := strings.SplitN(string(data), "\n", 2)
	if len(split) < 2 {
		err = NO_RESPONSE
		return
	}
	err = NewFail(split[0])

	log.Debugf("************  DUMP: %v:\n%s\n", err, string(data))

	switch {
	case err == nil:
		// TODO Stats.serverContact
		resp = split[1]
	case err == KEY_EXPIRED && retryHandler != nil:
		log.Warn("Server reported expired key; attempting to refresh time from server and retrying")

		retryHandler.RefreshServerStat()
		return retryHandler.client._Get(retryHandler.URL(retryAct, ""), "", nil)
	}
	return
}

type Fail string

func (fail Fail) Error() string {
	return fmt.Sprintf("FAIL: %s", string(fail))
}

var (
	NO_RESPONSE             = Fail("NO_RESPONSE")
	SERVER_ERROR            = Fail("SERVER_ERROR")
	TEMPORARILY_UNAVAILABLE = Fail("TEMPORARILY_UNAVAILABLE")

	// there really was a cake
	KEY_EXPIRED = Fail("KEY_EXPIRED")

	FAIL_CONNECT_TEST           = Fail("FAIL_CONNECT_TEST")
	FAIL_STARTUP_FLOOD          = Fail("FAIL_STARTUP_FLOOD")
	FAIL_OTHER_CLIENT_CONNECTED = Fail("FAIL_OTHER_CLIENT_CONNECTED")
	FAIL_CID_IN_USE             = Fail("FAIL_CID_IN_USE")
	FAIL_RESET_SUSPENDED        = Fail("FAIL_RESET_SUSPENDED")
)

func is(wat, prefix string) bool {
	return strings.HasPrefix(wat, prefix)
}

func NewFail(code string) (err error) {
	switch {
	case is(code, "OK"): // pass

	case is(code, "Log Code") || is(code, "Database Error"):
		return SERVER_ERROR
	case is(code, "TEMPORARILY_UNAVAILABLE"):
		return TEMPORARILY_UNAVAILABLE

	case is(code, "KEY_EXPIRED"):
		return KEY_EXPIRED

	case is(code, "FAIL_CONNECT_TEST"):
		return FAIL_CONNECT_TEST
	case is(code, "FAIL_STARTUP_FLOOD"):
		return FAIL_STARTUP_FLOOD
	case is(code, "FAIL_OTHER_CLIENT_CONNECTED"):
		return FAIL_OTHER_CLIENT_CONNECTED
	case is(code, "FAIL_CID_IN_USE"):
		return FAIL_CID_IN_USE
	case is(code, "FAIL_RESET_SUSPENDED"):
		return FAIL_RESET_SUSPENDED
	default:
		panic("undefined error " + code)
	}
	return
}
