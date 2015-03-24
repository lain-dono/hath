package main

import (
	"./log"
	"errors"
	"time"
)

var (
	FAIL_CONNECT_TEST           = errors.New("FAIL_CONNECT_TEST")
	FAIL_STARTUP_FLOOD          = errors.New("FAIL_STARTUP_FLOOD")
	FAIL_OTHER_CLIENT_CONNECTED = errors.New("FAIL_OTHER_CLIENT_CONNECTED")
	FAIL_CID_IN_USE             = errors.New("FAIL_CID_IN_USE")
	FAIL_RESET_SUSPENDED        = errors.New("FAIL_RESET_SUSPENDED")
)

func (api *API) notifyStart() (err error) {
	resp, err := api.Get(ACT_CLIENT_START /*, this*/)
	if err == nil {
		log.Info("Start notification successful. Note that there may be a short wait before the server registers this client on the network.")
		return nil
	}

	switch err := err.(type) {
	default:
		log.Error(err)
		return err
	case APIFail:
		switch {
		default:
			log.Error(err)
			return err
		case err.Is("FAIL_CONNECT_TEST"):
			log.Infof(FAIL_CONNECT_TEST_msg, Settings.getClientPort(), Settings.getClientHost())
			return FAIL_CONNECT_TEST

		case err.Is("FAIL_STARTUP_FLOOD"):
			log.Info(FAIL_STARTUP_FLOOD_msg)
			time.Sleep(90 * time.Second)
			return api.notifyStart()

		case err.Is("FAIL_OTHER_CLIENT_CONNECTED"):
			log.Info(FAIL_OTHER_CLIENT_CONNECTED_msg)
			//client.dieWithError(FAIL_OTHER_CLIENT_CONNECTED)
			return FAIL_OTHER_CLIENT_CONNECTED

		case err.Is("FAIL_CID_IN_USE"):
			log.Info(FAIL_CID_IN_USE_msg)
			//client.dieWithError(FAIL_CID_IN_USE)
			return FAIL_CID_IN_USE

		case err.Is("FAIL_RESET_SUSPENDED"):
			log.Info(FAIL_RESET_SUSPENDED_msg)
			//client.dieWithError(FAIL_RESET_SUSPENDED)
			return FAIL_RESET_SUSPENDED
		}
	}

	panic("WTF?")

	//return //false
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
