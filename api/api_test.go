package api

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	ts := &TestServer{}
	clientapi := httptest.NewServer(ts)
	defer clientapi.Close()

	Convey("All", t, func() {
		Convey("Fail", func() {
			//So(NewFail(SERVER_ERROR), ShouldEqual, SERVER_ERROR)
			//So(NewFail(FAIL_CONNECT_TEST.Error()), ShouldEqual, FAIL_CONNECT_TEST)
			//So(NewFail(FAIL_STARTUP_FLOOD.Error()), ShouldEqual, FAIL_STARTUP_FLOOD)
			//So(NewFail(FAIL_OTHER_CLIENT_CONNECTED.Error()), ShouldEqual, FAIL_OTHER_CLIENT_CONNECTED)
			//So(NewFail(FAIL_CID_IN_USE.Error()), ShouldEqual, FAIL_CID_IN_USE)
			//So(NewFail(FAIL_RESET_SUSPENDED.Error()), ShouldEqual, FAIL_RESET_SUSPENDED)

			//NO_RESPONSE             = Fail("NO_RESPONSE")
			//SERVER_ERROR            = Fail("SERVER_ERROR")
			//TEMPORARILY_UNAVAILABLE = Fail("TEMPORARILY_UNAVAILABLE")

			//// there really was a cake
			//KEY_EXPIRED = Fail("KEY_EXPIRED")
		})

		Convey("Client", func() {
			client := New(clientapi.URL)

			Convey("RefreshServerStat", func() {
				s, _ := client.RefreshServerStat()
				So(s, ShouldResemble, stat)
				ts.testAct(ACT_SERVER_STAT)
			})

			Convey("call Suspend", func() {
				ts.testNotify(client.Suspend(), ACT_CLIENT_SUSPEND)
			})
			Convey("call Resume", func() {
				ts.testNotify(client.Resume(), ACT_CLIENT_RESUME)
			})
			Convey("call Shutdown", func() {
				ts.testNotify(client.Shutdown(), ACT_CLIENT_STOP)
			})
			Convey("call MoreFiles", func() {
				ts.testNotify(client.MoreFiles(), ACT_MORE_FILES)
			})
			Convey("call Overload", func() {
				ts.testNotify(client.Overload(), ACT_OVERLOAD)
			})

			Convey("try login", func() {
				s, _ := client.LoadClientSettingsFromServer()
				So(s, ShouldResemble, settings)
				ts.testAct(ACT_CLIENT_LOGIN)
				So(client.IsLoginValidated(), ShouldBeTrue)
			})

			Convey("try start", func() {
				ts.testNotify(client.Start(), ACT_CLIENT_START)
			})
		})
	})
}

type TestServer struct {
	act string
}

func (ts *TestServer) testNotify(err error, act string) {
	So(err, ShouldBeNil)
	So(ts.act, ShouldEqual, act)
}
func (ts *TestServer) testAct(act string) {
	So(ts.act, ShouldEqual, act)
}

func (ts *TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ts.act = r.FormValue("act")
	switch ts.act {
	default:
		panic("fail act " + ts.act)

	case ACT_SERVER_STAT:
		fmt.Fprintln(w, "OK")
		fmt.Fprintln(w)
		fmt.Fprintln(w, strings.Join(stat, "\n"))
	case ACT_CLIENT_SUSPEND, ACT_CLIENT_RESUME, ACT_CLIENT_STOP, ACT_MORE_FILES, ACT_OVERLOAD:
		fmt.Fprintln(w, "OK")
	case ACT_CLIENT_LOGIN:
		fmt.Fprintln(w, "OK")
		fmt.Fprintln(w, strings.Join(settings, "\n"))
	case ACT_CLIENT_START:
		fmt.Fprintln(w, "OK")
	}
}

const (
	server_time = "1400000000"
)

var stat = []string{
	"min_client_build=85",
	"cur_client_build=88",
	"server_time=" + server_time,
}

var settings = []string{
	"rpc_server_ip=::ffff:0.0.0.0;::ffff:10.0.0.0",
	"image_server=ul.e-hentai.org",
	"name=lo",
	"host=::ffff:127.0.0.1",
	"port=1024",
	"throttle_bytes=8200000",
	"hourbwlimit_bytes=0",
	"disklimit_bytes=53687091200",
	"diskremaining_bytes=0",
	"request_server=g.e-hentai.org",
	"request_proxy_mode=1",
	"disable_bwm=true",
	"static_ranges=000a;00fa;0114;",
}
