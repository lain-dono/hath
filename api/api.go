package main

import (
	"./cache"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	ACT_SERVER_STAT     = "server_stat"
	ACT_GET_BLACKLIST   = "get_blacklist"
	ACT_CLIENT_LOGIN    = "client_login"
	ACT_CLIENT_SETTINGS = "client_settings"
	ACT_CLIENT_START    = "client_start"
	ACT_CLIENT_SUSPEND  = "client_suspend" // notifiy
	ACT_CLIENT_RESUME   = "client_resume"  // notifiy
	ACT_CLIENT_STOP     = "client_stop"    // notifiy
	ACT_STILL_ALIVE     = "still_alive"
	ACT_FILE_UNCACHE    = "file_uncache"  // notifiy
	ACT_FILE_REGISTER   = "file_register" // notifiy
	ACT_MORE_FILES      = "more_files"    // notifiy
	ACT_FILE_TOKENS     = "download_list"
	ACT_OVERLOAD        = "overload" // notifiy
)

type API struct {
	Id  int64
	Key string

	serverTimeDelta time.Duration

	lastOverloadNotification time.Time
	loginValidated           bool
	*http.Client
}

func NewAPI() *API {
	tr := &http.Transport{}
	return &API{
		Client: &http.Client{
			Timeout:   3600 * time.Second, //3600 000
			Transport: tr,
		},
	}
}
func (api *API) IsLoginValidated() bool { return api.loginValidated }

// communications that do not use additional variables can use this
func (api *API) notifiySimple(act, humanReadable string) error {
	return api.notifiy(act, "", humanReadable)
}
func (api *API) notifiy(act, add, humanReadable string) error {
	_, err := api._Get(api.URL(act, add), "", api)
	if err != nil {
		log.Debugf("%s notification successful.", humanReadable)
	} else {
		log.Warn("%s notification failed: %s", humanReadable, err)
	}
	return err
}

// simple notifications
func (api *API) Suspend() error   { return api.notifiySimple(ACT_CLIENT_SUSPEND, "Suspend") }
func (api *API) Resume() error    { return api.notifiySimple(ACT_CLIENT_RESUME, "Resume") }
func (api *API) Shutdown() error  { return api.notifiySimple(ACT_CLIENT_STOP, "Shutdown") }
func (api *API) MoreFiles() error { return api.notifiySimple(ACT_MORE_FILES, "More Files") }
func (api *API) Overload() error {
	//< nowtime - 30000)
	now := time.Now()
	if now.After(api.lastOverloadNotification.Add(30 * time.Second)) {
		api.lastOverloadNotification = now
		return api.notifiySimple(ACT_OVERLOAD, "Overload")
	}
	return nil
}

// these communcation methods are more complex, and have their own result parsing

func (api *API) UncachedFiles(deletedFiles []cache.HVFile) {
	// NOTE: as we want to avoid POST, we do this as a long GET.
	// to avoid exceeding certain URL length limitations, we uncache at most 50 files at a time

	deleteCount := len(deletedFiles)
	next := func() (ret []cache.HVFile) {
		max := 50
		if max > len(deletedFiles) {
			max = len(deletedFiles)
		}

		ret, deletedFiles = deletedFiles[:max], deletedFiles[max:]
		deleteCount -= len(ret)
		return ret
	}

	if deleteCount != 0 {
		log.Debugf("Notifying server of %d uncached files...", deleteCount)

		for rm := next(); len(rm) != 0; rm = next() {
			ids := strings.Join(cache.HVFIds(rm), ";")

			_ = api.notifiy(ACT_FILE_UNCACHE, ids, "Uncache")
			// TODO
		}
	}
}

func (api *API) RegisterFiles(pendingRegister []cache.HVFile) error {
	log.Debug("Notifying server of %d registered files...", len(pendingRegister))
	ids := strings.Join(cache.HVFIds(pendingRegister), ";")
	return api.notifiy(ACT_FILE_REGISTER, ids, "Register")
}

func (api *API) stringList(act, add string) (list []string, err error) {
	resp, err := api.Get(api.URL(act, add))
	if err == nil {
		list = trim(strings.Split(resp.Text, "\n"))
	}
	return
}

func (api *API) Blacklist(delta time.Duration) (list []string, err error) {
	return api.stringList(ACT_GET_BLACKLIST, strconv.FormatInt(int64(delta.Seconds()), 0))
}

/*
	public void stillAliveTest() {
		CakeSphere cs = new CakeSphere(this, client);
		cs.stillAlive();
	}
*/

// this MUST NOT be called after the client has started up,
// as it will clear out and reset the client on the server,
// leaving the client in a limbo until restart
func (api *API) loadClientSettingsFromServer() (settings []string, err error) {
	//TODO Stats.setProgramStatus("Loading settings from server...");
	//log.Infof("Connecting to the Hentai@Home Server to register client with ID %d...", api.Id)

	log.Info("Reading Hentai@Home client settings from server...")
	resp, err := api.Get(api.URL(ACT_CLIENT_LOGIN, ""))
	switch e := err.(type) {
	case nil:
		api.loginValidated = true
		settings = trim(strings.Split(resp.Text, "\n"))
	case APIFail:
		log.Warnf("\nAuthentication failed, please re-enter your Client ID and Key (Code: %s)", e.Code)
		//TODO Settings.promptForIDAndKey(client.getInputQueryHandler());
	default:
		switch e {
		case NO_RESPONSE, SERVER_ERROR, TEMPORARILY_UNAVAILABLE:
			//log.Error("Failed to get a login response from server.", err)
			panic("Failed to get a login response from server.")
		default:
			panic(e)
		}
	}
	//}
	return
}

func (api *API) refreshServerSettings() (settings []string, err error) {
	log.Info("Refreshing Hentai@Home client settings from server...")
	return api.stringList(ACT_CLIENT_SETTINGS, "")
	// XXX wtf?
	//- we're not bothering to recheck the free space as the client doesn't accept live reductions of disk space
	//client.getCacheHandler().recheckFreeDiskSpace();
}

func trim(from []string) (to []string) {
	for _, v := range from {
		v = strings.TrimSpace(v)
		if v != "" {
			to = append(to, v)
		}
	}
	return
}

// get timestamp and minimum client build from server
func (api *API) refreshServerStat() (stat []string, err error) {
	// TODO Stats.setProgramStatus("Getting initial stats from server...");
	log.Info("Getting initial stats from server...")
	return api.stringList(ACT_SERVER_STAT, "")
}

func (api *API) FileTokens(requestTokens []string) (tokenTable map[string]string, err error) {
	tokens := strings.Join(requestTokens, ";") + ";"
	resp, err := api.Get(api.URL(ACT_FILE_TOKENS, tokens))

	if err != nil {
		log.Info("Could not grab token list - most likely the client has not been qualified yet. Will retry in a few minutes.")
		return
	}

	tokenTable = make(map[string]string)
	split := strings.Split(resp.Text, "\n")
	for _, s := range split {
		if s != "" {
			kv := strings.Split(s, " ")
			tokenTable[kv[0]] = kv[1]
		}
	}
	return
}
