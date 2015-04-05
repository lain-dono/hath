package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	//"net/http"
	"strings"
	"time"
)

const (
	CLIENT_BUILD   = 96
	CLIENT_VERSION = "1.2.5"
	CLIENT_API_URL = "http://rpc.hentaiathome.net/clientapi.php?"

	CLIENT_KEY_LENGTH = 20
)

func (api *API) ServerTime() int64 {
	t := time.Now().Add(api.serverTimeDelta)
	return t.Unix() //*1000 + t.Nanosecond()/(1000*1000)
}
func (api *API) SetServerTime(unix int64) {
	api.serverTimeDelta = time.Unix(unix, 0).Sub(time.Now())
	log.Debugf("Setting altered: serverTimeDelta=%s", api.serverTimeDelta)
}

func (api *API) URL(act, add string) (ret string) {
	defer func() { log.Debug(ret) }()
	if act == ACT_SERVER_STAT {
		return CLIENT_API_URL + fmt.Sprintf("clientbuild=%d&act=%s", CLIENT_BUILD, act)
	}

	correctedTime := api.ServerTime()
	actkey := fmt.Sprintf("hentai@home-%s-%s-%d-%d-%s",
		act, add, api.Id, correctedTime, api.Key)
	href := fmt.Sprintf("clientbuild=%d&act=%s&add=%s&cid=%d&acttime=%d&actkey=%s",
		CLIENT_BUILD, act, add, api.Id, correctedTime, SHA(actkey))

	return CLIENT_API_URL + href
}

func (api *API) Get(url string) (resp APIResponse, err error) {
	return api._Get(url, "", api)
}

func (api *API) _Get(url, retryAct string, retryHandler *API) (resp APIResponse, err error) {
	r, err := api.Client.Get(url)
	if err != nil {
		log.Error("xx", url)
		return
	}
	defer r.Body.Close()

	resp.Data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	split := strings.SplitN(string(resp.Data), "\n", 2)

	switch {
	case len(split) == 0:
		err = NO_RESPONSE
	case strings.HasPrefix(split[0], "Log Code") || strings.HasPrefix(split[0], "Database Error"):
		err = SERVER_ERROR
	case strings.HasPrefix(split[0], "TEMPORARILY_UNAVAILABLE"):
		err = TEMPORARILY_UNAVAILABLE
	case split[0] == "OK":
		// TODO Stats.serverContact
		resp.Text = split[1]
	case split[0] == "KEY_EXPIRED" && retryHandler != nil:
		log.Warn("Server reported expired key; attempting to refresh time from server and retrying")

		retryHandler.refreshServerStat()
		return retryHandler._Get(retryHandler.URL(retryAct, ""), "", nil)
		panic("not implemented")
	default:
		err = APIFail{split[0], split[1]}
	}
	return
}

///

var (
	NO_RESPONSE             = errors.New("NO_RESPONSE")
	SERVER_ERROR            = errors.New("SERVER_ERROR")
	TEMPORARILY_UNAVAILABLE = errors.New("TEMPORARILY_UNAVAILABLE")
)

type APIFail struct {
	Code string
	Text string
}

func (fail APIFail) Error() string {
	return fmt.Sprintf("FAIL: %s: %s", fail.Code, fail.Text)
}
func (fail APIFail) Is(prefix string) bool {
	return strings.HasPrefix(fail.Code, prefix)
}

type APIResponse struct {
	Text string
	Data []byte
}

func (resp *APIResponse) String() string {
	return fmt.Sprintf("APIResponse {Text=%s}", resp.Text)
}
