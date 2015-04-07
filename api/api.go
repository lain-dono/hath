package api

import (
	"../cache"
	"../util"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TODO RefreshServerStat parse not in settings.go

var log = util.Logger()

const (
	ServerStatQuery = "?clientbuild=%d&act=%s"
	ApiQuery        = "?clientbuild=%d&act=%s&add=%s&cid=%d&acttime=%d&actkey=%s"
	ActKeyFormat    = "hentai@home-%s-%s-%d-%d-%s"
)

const (
	CLIENT_BUILD   = 96
	CLIENT_VERSION = "1.2.5"
	CLIENT_API_URL = "http://rpc.hentaiathome.net/clientapi.php"

	CLIENT_KEY_LENGTH = 20

	MAX_KEY_TIME_DRIFT = 300
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
	Login
	Time

	base string

	lastOverloadNotification time.Time

	client *Client
}

func New(base string) (api *API) {
	return &API{
		base:   base,
		client: newClient(),
	}
}

// communications that do not use additional variables can use this

func (api *API) notify(act, add string) error {
	errc := make(chan error)
	defer close(errc)
	go func() {
		_, err := api.client._Get(api.URL(act, add), "", api)
		if err == nil {
			log.Debugf("%s notification successful.", strings.ToUpper(act))
		} else {
			log.Warnf("%s notification failed: %s", strings.ToUpper(act), err)
		}
		errc <- err
	}()
	return <-errc
}

func (api *API) get(act, add string) (list []string, err error) {
	errc := make(chan error)
	defer close(errc)
	go func() {
		resp, err := api.client._Get(api.URL(act, add), "", api)
		if err == nil {
			list = trim(strings.Split(resp, "\n"))
			log.Debugf("%s get successful. len: %d", strings.ToUpper(act), len(list))
		} else {
			log.Warnf("%s get failed: %s", strings.ToUpper(act), err)
		}
		errc <- err
	}()
	return list, <-errc
}

func (api *API) URL(act, add string) (ret string) {
	defer func() { log.Debug(ret) }()
	if act == ACT_SERVER_STAT {
		return api.base + fmt.Sprintf(ServerStatQuery, CLIENT_BUILD, act)
	}

	correctedTime := api.ServerTime()
	actkey := fmt.Sprintf(ActKeyFormat, act, add, api.Id, correctedTime, api.Key)
	href := fmt.Sprintf(ApiQuery, CLIENT_BUILD, act, add, api.Id, correctedTime, util.SHA(actkey))

	return api.base + href
}

// simple notifications
func (api *API) Suspend() error   { return api.notify(ACT_CLIENT_SUSPEND, "") }
func (api *API) Resume() error    { return api.notify(ACT_CLIENT_RESUME, "") }
func (api *API) Shutdown() error  { return api.notify(ACT_CLIENT_STOP, "") }
func (api *API) MoreFiles() error { return api.notify(ACT_MORE_FILES, "") }
func (api *API) Overload() error {
	//< nowtime - 30000)
	now := time.Now()
	if now.After(api.lastOverloadNotification.Add(30 * time.Second)) {
		api.lastOverloadNotification = now
		return api.notify(ACT_OVERLOAD, "")
	}
	return nil
}

// these communcation methods are more complex, and have their own result parsing

func (api *API) UncachedFiles(deletedFiles []cache.Id) {
	// NOTE: as we want to avoid POST, we do this as a long GET.
	// to avoid exceeding certain URL length limitations, we uncache at most 50 files at a time

	next := func(max int) (ret []cache.Id) {
		if max > len(deletedFiles) {
			max = len(deletedFiles)
		}
		ret, deletedFiles = deletedFiles[:max], deletedFiles[max:]
		return ret
	}

	log.Debugf("Notifying server of %d uncached files...", len(deletedFiles))

	for rm := next(50); len(rm) != 0; rm = next(50) {
		ids := strings.Join(cache.Ids(rm), ";")

		_ = api.notify(ACT_FILE_UNCACHE, ids)
		// TODO
	}
}

func (api *API) RegisterFiles(pendingRegister []cache.Id) error {
	log.Debugf("Notifying server of %d registered files...", len(pendingRegister))
	ids := strings.Join(cache.Ids(pendingRegister), ";")
	return api.notify(ACT_FILE_REGISTER, ids)
}

func (api *API) Blacklist(delta time.Duration) (list []string, err error) {
	// TODO
	return api.get(ACT_GET_BLACKLIST, strconv.FormatInt(int64(delta.Seconds()), 0))
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
func (api *API) LoadClientSettingsFromServer() (settings []string, err error) {
	//TODO Stats.setProgramStatus("Loading settings from server...");
	//log.Infof("Connecting to the Hentai@Home Server to register client with ID %d...", api.Id)

	log.Info("Reading Hentai@Home client settings from server...")
	resp, err := api.get(ACT_CLIENT_LOGIN, "")
	switch e := err.(type) {
	case nil:
		api.loginValidated = true
		settings = resp
	case Fail:
		log.Warnf("\nAuthentication failed, please re-enter your Client ID and Key (Code: %s)", e)
		//TODO Settings.promptForIDAndKey(client.getInputQueryHandler());
		switch e {
		case NO_RESPONSE, SERVER_ERROR, TEMPORARILY_UNAVAILABLE:
			//log.Error("Failed to get a login response from server.", err)
			panic("Failed to get a login response from server.")
		default:
			panic(e)
		}
	default:
		panic(e)
	}
	//}
	return
}

func (api *API) RefreshServerSettings() (settings []string, err error) {
	log.Info("Refreshing Hentai@Home client settings from server...")
	return api.get(ACT_CLIENT_SETTINGS, "")
	// XXX wtf?
	//- we're not bothering to recheck the free space as the client doesn't accept live reductions of disk space
	//client.getCacheHandler().recheckFreeDiskSpace();
}

// get timestamp and minimum client build from server
func (api *API) RefreshServerStat() (stat []string, err error) {
	// TODO Stats.setProgramStatus("Getting initial stats from server...");
	log.Info("Getting initial stats from server...")
	return api.get(ACT_SERVER_STAT, "")
}

func (api *API) FileTokens(requestTokens []string) (tokenTable map[string]string, err error) {
	tokens := strings.Join(requestTokens, ";") + ";"
	resp, err := api.get(ACT_FILE_TOKENS, tokens)

	if err != nil {
		log.Info("Could not grab token list - most likely the client has not been qualified yet. Will retry in a few minutes.")
		return
	}

	tokenTable = make(map[string]string)
	for _, s := range resp {
		if s != "" {
			kv := strings.Split(s, " ")
			tokenTable[kv[0]] = kv[1]
		}
	}
	return
}

func (api *API) Start() (err error) {
	err = api.notify(ACT_CLIENT_START, "")

	switch err {
	default:
		log.Error(err)
	case nil:
		log.Info("Start notification successful. Note that there may be a short wait before the server registers this client on the network.")
	case FAIL_STARTUP_FLOOD:
		log.Info(FAIL_STARTUP_FLOOD_msg)
		time.Sleep(90 * time.Second)
		return api.Start()

	case FAIL_CONNECT_TEST:
		log.Infof(FAIL_CONNECT_TEST_msg) //TODO , Settings.getClientPort(), Settings.getClientHost())
	case FAIL_OTHER_CLIENT_CONNECTED:
		log.Info(FAIL_OTHER_CLIENT_CONNECTED_msg)
		//client.dieWithError(FAIL_OTHER_CLIENT_CONNECTED)
	case FAIL_CID_IN_USE:
		log.Info(FAIL_CID_IN_USE_msg)
		//client.dieWithError(FAIL_CID_IN_USE)
	case FAIL_RESET_SUSPENDED:
		log.Info(FAIL_RESET_SUSPENDED_msg)
		//client.dieWithError(FAIL_RESET_SUSPENDED)
	}

	return err
}

const FAIL_CONNECT_TEST_msg = `
************************************************************************************************************************************
The client has failed the external connection test.
The server failed to verify that this client is online and available from the Internet.
If you are behind a firewall, please check that port %s is forwarded to this computer.
You might also want to check that %s is your actual public IP address.
If you need assistance with forwarding a port to this client, locate a guide for your particular router at http://portforward.com/
The client will remain running so you can run port connection tests.
Use Program -> Exit in windowed mode or hit Ctrl+C in console mode to exit the program.
************************************************************************************************************************************
`
const FAIL_STARTUP_FLOOD_msg = `
************************************************************************************************************************************
Flood control is in effect.
The client will automatically retry connecting in 90 seconds.
************************************************************************************************************************************
`
const FAIL_OTHER_CLIENT_CONNECTED_msg = `
************************************************************************************************************************************
"The server detected that another client was already connected from this computer or local network.
You can only have one client running per public IP address.
The program will now terminate.
************************************************************************************************************************************
`
const FAIL_CID_IN_USE_msg = `
************************************************************************************************************************************
The server detected that another client is already using this client ident.
If you want to run more than one client, you have to apply for additional idents.
The program will now terminate.
************************************************************************************************************************************
`
const FAIL_RESET_SUSPENDED_msg = `
************************************************************************************************************************************
This client ident has been revoked for having too many cache resets.
The program will now terminate.
************************************************************************************************************************************
`
