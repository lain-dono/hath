package main

import (
	//"net/url"
	"./cache"
	"./log"
	//"errors"
	//"fmt"
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
	ACT_CLIENT_SUSPEND  = "client_suspend"
	ACT_CLIENT_RESUME   = "client_resume"
	ACT_CLIENT_STOP     = "client_stop"
	ACT_STILL_ALIVE     = "still_alive"
	ACT_FILE_UNCACHE    = "file_uncache"
	ACT_FILE_REGISTER   = "file_register"
	ACT_MORE_FILES      = "more_files"
	ACT_FILE_TOKENS     = "download_list"
	ACT_OVERLOAD        = "overload"
)

type API struct {
	Id  int
	Key string

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
func (api *API) simple(act, humanReadable string) error {
	_, err := api.Get(api.URL(act, humanReadable))
	if err != nil {
		log.Debug(humanReadable + " notification successful.")
	} else {
		log.Warn(humanReadable + " notification failed.")
	}
	return err
}

// simple notifications
func (api *API) Suspend() error {
	return api.simple(ACT_CLIENT_SUSPEND, "Suspend")
}
func (api *API) Resume() error {
	return api.simple(ACT_CLIENT_RESUME, "Resume")
}
func (api *API) Shutdown() error {
	return api.simple(ACT_CLIENT_STOP, "Shutdown")
}
func (api *API) notifyOverload() error {
	//< nowtime - 30000)
	now := time.Now()
	if now.After(api.lastOverloadNotification.Add(30 * time.Second)) {
		api.lastOverloadNotification = now
		return api.simple(ACT_OVERLOAD, "Overload")
	}
	return nil
}
func (api *API) MoreFiles() error {
	return api.simple(ACT_MORE_FILES, "More Files")
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
			addr := api.URL(ACT_FILE_UNCACHE, ids)
			_, err := api.Get(addr)
			if err == nil {
				log.Debug("Uncache notification successful.")
			} else {
				log.Warn("Uncache notification failed.", err)
			}
		}
	}
}

func (api *API) RegisterFiles(pendingRegister []cache.HVFile) {
	log.Debug("Notifying server of %d registered files...", len(pendingRegister))

	ids := strings.Join(cache.HVFIds(pendingRegister), ";")

	addr := api.URL(ACT_FILE_UNCACHE, ids)
	_, err := api.Get(addr)
	if err == nil {
		log.Debug("Register notification successful.")
	} else {
		log.Warn("Register notification failed.", err)
	}
}

func (api *API) Blacklist(delta time.Duration) (list []string, err error) {
	addr := api.URL(ACT_GET_BLACKLIST, strconv.FormatInt(int64(delta.Seconds()), 0))
	resp, err := api.Get(addr)
	if err == nil {
		list = strings.Split(resp.Text, "\n")
	}
	return
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
func (api *API) loadClientSettingsFromServer() (err error) {
	//TODO Stats.setProgramStatus("Loading settings from server...");
	log.Infof("Connecting to the Hentai@Home Server to register client with ID %d...", api.Id)

	addr := api.URL(ACT_CLIENT_LOGIN, "")

	for {
		err = api.refreshServerStat()
		if err != nil {
			//HentaiAtHomeClient.dieWithError("Failed to get initial stat from server.");
			return
		}

		log.Info("Reading Hentai@Home client settings from server...")
		resp, err := api.Get(addr)
		switch e := err.(type) {
		case nil:
			api.loginValidated = true
			log.Info("Applying settings...")
			//TODO Settings.parseAndUpdateSettings(sr.getResponseText());
			log.Info("Finished applying settings")
			break // XXX

		case APIFail:
			log.Warn("\nAuthentication failed, please re-enter your Client ID and Key (Code: %s)", e.Code)
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
	}
}

func (api *API) refreshServerSettings() (err error) {
	log.Info("Refreshing Hentai@Home client settings from server...")

	resp, err := api.Get(api.URL(ACT_CLIENT_SETTINGS, ""))
	if err == nil {
		//TODO Settings.parseAndUpdateSettings(sr.getResponseText());
		log.Info("Finished applying settings")
		//- we're not bothering to recheck the free space as the client doesn't accept live reductions of disk space
		//client.getCacheHandler().recheckFreeDiskSpace();
	}
	log.Warn("Failed to refresh settings")
	return
}

func (api *API) refreshServerStat() (err error) {
	// TODO Stats.setProgramStatus("Getting initial stats from server...");
	log.Info("Getting initial stats from server...")

	// get timestamp and minimum client build from server
	resp, err := api.Get(api.URL(ACT_SERVER_STAT, ""))

	if err == nil {
		//TODO Settings.parseAndUpdateSettings(sr.getResponseText());
	}
	return
}

func (api *API) getFileTokens(requestTokens []string) (tokenTable map[string]string, err error) {
	tokens := strings.Join(requestTokens, ";") + ";"

	resp, err := api.Get(api.URL(ACT_FILE_TOKENS, ""))

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
