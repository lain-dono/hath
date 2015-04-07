package main

import (
	"./api"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
)

const (
	MAX_CONNECTION_BASE = 20
)

func promptForIDAndKey() (login api.Login) {
	log.Info(`
	Before you can use this client, you will have to register it at http://hentaiathome.net/
	IMPORTANT: YOU NEED A SEPARATE IDENT FOR EACH CLIENT YOU WANT TO RUN.
	DO NOT ENTER AN IDENT THAT WAS ASSIGNED FOR A DIFFERENT CLIENT.");
	After registering, enter your ID and Key below to start your clientapi..
	(You will only have to do this once.)
`)

	for {
		fmt.Print("Enter Client ID: ")
		var val string
		fmt.Scanln(&val)
		login.Id, _ = strconv.ParseInt(val, 10, 64)

		if login.CheckId() {
			break
		}
		log.Warn("Invalid Client ID. Please try again.")
	}

	for {
		fmt.Print("Enter Client Key: ")
		fmt.Scanln(&login.Key)

		if login.CheckKey() {
			break
		}
		log.Warn("Invalid Client Key, it must be exactly 20 alphanumerical characters. Please try again.")
	}

	return
}

//DATA_FILENAME_CLIENT_LOGIN
func loadClientLoginFromFile(name string) (login api.Login, err error) {
	defer func() {
		if err != nil || !login.CheckId() || !login.CheckKey() {
			log.Warn("Encountered error when reading %s: %s", name, err)
		}
	}()
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return
	}
	split := strings.SplitN(string(b), "-", 2)
	if len(split) == 2 {
		login.Key = split[1]
		login.Id, err = strconv.ParseInt(split[0], 0, 0)
	}
	return
}

func parseAndUpdateSettings(settings []string) (err error) {
	log.Info("Applying settings...")
	log.Debug("Settings: ", settings)
	for _, s := range settings {
		split := strings.SplitN(s, "=", 2)
		updateSetting(strings.ToLower(split[0]), split[1])
	}
	log.Info("Finished applying settings")

	return
}

//var Settings = &settings{}

func updateSetting(setting, value string) bool {
	setting = strings.ToLower(strings.Replace(setting, "-", "_", -1))
	i, _ := strconv.ParseInt(value, 10, 0)

	switch setting {
	case "min_client_build":
		if i > api.CLIENT_BUILD {
			panic("Your client is too old to connect to the Hentai@Home Network. Please download the new version of the client from http://hentaiathome.net/")
		}
	case "cur_client_build":
		if i > api.CLIENT_BUILD {
			//warnNewClient = true
			log.Warn("A new client version is available. Please download it from http://hentaiathome.net/ at your convenience.")
		}
	case "server_time":
		Client.SetServerTime(i)
		return true

	case "rpc_server_ip": // TODO
		split := strings.Split(value, ";")
		for _, s := range split {
			log.Infof("'%s': %s", s, net.ParseIP(s))
			//rpcServers = append
		}
	case "image_server":
		//imageServer = value
		log.Printf("image_server: %s", value)
	case "name":
		//clientName = value
		log.Printf("name: %s", value)
	case "host":
		//clientHost = value
		log.Printf("host: %s", value)
	case "port":
		//clientPort = i
		log.Printf("port: %d", i)
	case "request_server":
		//requestServer = value
		log.Printf("request_server: %s", value)
	case "request_proxy_mode":
		//requestProxyMode = i
		log.Printf("request_proxy_mode: %d", i)
	case "throttle_bytes":
		// THIS SHOULD NOT BE ALTERED BY THE CLIENT AFTER STARTUP.
		// Using the website interface will update the throttle value for the dispatcher first,
		// and update the client on the first stillAlive test.
		//throttle_bytes = i
		log.Printf("throttle_bytes: %d", i)
	case "hourbwlimit_bytes":
		// see above
		//hourbwlimit_bytes = i
		log.Printf("hourbwlimit_bytes: %d", i)

	case "disklimit_bytes":
		newLimit := i
		log.Printf("disklimit_bytes: %d", newLimit)
		if newLimit >= CacheHandler.DiskLimitBytes {
			CacheHandler.DiskLimitBytes = newLimit
		} else {
			log.Warn("The disk limit has been reduced. However, this change will not take effect until you restart your client.")
		}

	case "diskremaining_bytes":
		//diskremaining_bytes = i
		log.Printf("diskremaining_bytes: %d", i)
	case "force_dirty":
		//forceDirty = value == "true"
		log.Printf("force_dirty: %s", value)
	case "verify_cache":
		//verifyCache = value == "true"
		log.Printf("verify_cache: %s", value)
	case "use_less_memory":
		//useLessMemory = value == "true"
		log.Printf("use_less_memory: %s", value)
	case "disable_logging":
		//Out.disableLogging()
		log.Printf("disable_logging")
	case "disable_bwm":
		//disableBWM = value == "true"
		log.Printf("disable_logging")
	case "skip_free_space_check":
		//skipFreeSpaceCheck = value == "true"
		log.Printf("skip_free_space_check %s", value)
	case "max_connections":
		//overrideConns = i
		log.Printf("max_connections %d", i)

	case "static_ranges": // TODO
		staticRanges := make(map[string]bool)

		split := strings.Split(value, ";")
		for _, s := range split {
			if len(s) == 4 {
				staticRanges[s] = true
			}
		}
		log.Info("staticRanges: ", staticRanges)
		CacheHandler.SetStaticRanges(staticRanges)
		log.Info("staticRanges end")

	case "silentstart":
		// pass
		log.Printf("SILENTSTART")

	default:
		// don't flag errors if the setting is handled by the GUI
		log.Warn("Unknown setting " + setting + " = " + value)
		return false
	}

	//Out.debug("Setting altered: " + setting +"=" + value);
	//return true;
	//} catch(Exception e) {
	//Out.warning("Failed parsing setting " + setting + " = " + value);
	//}
	//*/

	return false
}

/////////////////////////

// +build ignore

/*

Copyright 2008-2014 E-Hentai.org
http://forums.e-hentai.org/
ehentai@gmail.com

This file is part of Hentai@Home.

Hentai@Home is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Hentai@Home is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Hentai@Home.  If not, see <http://www.gnu.org/licenses/>.

*/
/*
var (
	// the client build is among others used by the server to determine the client's capabilities.
	// any forks should use the build number as an indication of compatibility with mainline,
	// and not use it as an internal build number.

	CLIENT_BUILD   = 96
	CLIENT_VERSION = "1.2.5"

	CLIENT_API_URL = "http://rpc.hentaiathome.net/clientapi.php?"

	DATA_FILENAME_CLIENT_LOGIN    = "client_login"
	DATA_FILENAME_LASTHIT_HISTORY = "lasthit_history"

	CLIENT_KEY_LENGTH   = 20
	MAX_KEY_TIME_DRIFT  = 300
	MAX_CONNECTION_BASE = 20

	CONTENT_TYPE_DEFAULT = "text/html; charset=iso-8859-1"
	CONTENT_TYPE_OCTET   = "application/octet-stream"
	CONTENT_TYPE_JPG     = "image/jpeg"
	CONTENT_TYPE_PNG     = "image/png"
	CONTENT_TYPE_GIF     = "image/gif"

	TCP_PACKET_SIZE_HIGH = 1460
	TCP_PACKET_SIZE_LOW  = 536
	MAX_REQUEST_LENGTH   = 10000
)*/

/*
package org.hath.base;

import java.io.File;
import java.net.InetAddress;
import java.util.Hashtable;

public class Settings {
	public static final String NEWLINE = System.getProperty("line.separator");

	private static HentaiAtHomeClient activeClient;
	private static HathGUI activeGUI;

	private static int clientID = 0;
	private static String clientKey = "";
	private static int serverTimeDelta = 0;

	private static Object rpcChangeMonitor = new Object();
	private static InetAddress rpcServers[] = null;
	private static String imageServer = "";
	private static String clientName = "";
	private static String clientHost = "";
	private static int clientPort = 0;
	private static String requestServer = "";
	private static int requestProxyMode = 0;

	private static int throttle_bytes = 0;
	private static long hourbwlimit_bytes = 0;
	private static long disklimit_bytes = 0;
	private static long diskremaining_bytes = 0;

	// read from command-line arguments

	private static boolean forceDirty = false;
	private static boolean verifyCache = false;
	private static boolean skipFreeSpaceCheck = false;
	private static boolean warnNewClient = false;
	private static boolean useLessMemory = false;
	private static boolean disableBWM = false;

	private static int overrideConns = 0;

	private static File datadir = null;

	private static Hashtable<String, Integer> staticRanges = null;

	public static void setActiveClient(HentaiAtHomeClient client) {
		activeClient = client;
	}

	public static void setActiveGUI(HathGUI gui) {
		activeGUI = gui;
	}

	public static boolean loginCredentialsAreSyntaxValid() {
		return clientID > 0 && java.util.regex.Pattern.matches("^[a-zA-Z0-9]{" + Settings.CLIENT_KEY_LENGTH + "}$", clientKey);
	}

	public static boolean loadClientLoginFromFile() {
		File clientLogin = new File(Settings.getDataDir(), Settings.DATA_FILENAME_CLIENT_LOGIN);

		if(!clientLogin.exists()) {
			return false;
		}

		try {
			String filecontent = FileTools.getStringFileContents(clientLogin);

			if(!filecontent.isEmpty()) {
				String[] split = filecontent.split("-", 2);

				if(split.length == 2) {
					clientID = Integer.parseInt(split[0]);
					clientKey = split[1];

					return true;
				}
			}
		} catch(Exception e) {
			Out.warning("Encountered error when reading " + Settings.DATA_FILENAME_CLIENT_LOGIN + ": " + e);
		}

		return false;
	}

	public static void promptForIDAndKey(InputQueryHandler iqh) {
		Out.info("Before you can use this client, you will have to register it at http://hentaiathome.net/");
		Out.info("IMPORTANT: YOU NEED A SEPARATE IDENT FOR EACH CLIENT YOU WANT TO RUN.");
		Out.info("DO NOT ENTER AN IDENT THAT WAS ASSIGNED FOR A DIFFERENT CLIENT.");
		Out.info("After registering, enter your ID and Key below to start your client.");
		Out.info("(You will only have to do this once.)\n");

		clientID = 0;
		clientKey = "";

		do {
			try {
				clientID = Integer.parseInt(iqh.queryString("Enter Client ID").trim());
			} catch(java.lang.NumberFormatException nfe) {
				Out.warning("Invalid Client ID. Please try again.");
			}
		} while(clientID < 1000);

		do {
			clientKey = iqh.queryString("Enter Client Key").trim();
			if(!loginCredentialsAreSyntaxValid()) {
				Out.warning("Invalid Client Key, it must be exactly 20 alphanumerical characters. Please try again.");
			}
		} while(!loginCredentialsAreSyntaxValid());

		try {
			FileTools.putStringFileContents(new File(Settings.getDataDir(), Settings.DATA_FILENAME_CLIENT_LOGIN), clientID + "-" + clientKey);
		} catch(java.io.IOException ioe) {
			Out.warning("Error encountered when writing " + Settings.DATA_FILENAME_CLIENT_LOGIN + ": " + ioe);
		}
	}

	public static boolean parseAndUpdateSettings(String[] settings) {
		if(settings == null) {
			return false;
		}

		for(String s : settings) {
			if(s != null) {
				String[] split = s.split("=", 2);

				if(split.length == 2) {
					updateSetting(split[0].toLowerCase(), split[1]);
				}
			}
		}

		return true;
	}

	// note that these settings will currently be overwritten by any equal ones read from the server, so it should not be used to override server-side settings.
	public static boolean parseArgs(String[] args) {
		if(args == null) {
			return false;
		}

		for(String s : args) {
			if(s != null) {
				if(s.startsWith("--")) {
					String[] split = s.substring(2).split("=", 2);

					if(split.length == 2) {
						updateSetting(split[0].toLowerCase(), split[1]);
					}
					else {
						updateSetting(split[0].toLowerCase(), "true");
					}
				}
				else {
					Out.warning("Invalid command argument: " + s);
				}
			}
		}

		return true;
	}

	public static boolean updateSetting(String setting, String value) {
		setting = setting.replace("-", "_");

		try {
			if(setting.equals("min_client_build")) {
				if(Integer.parseInt(value) > CLIENT_BUILD) {
					HentaiAtHomeClient.dieWithError("Your client is too old to connect to the Hentai@Home Network. Please download the new version of the client from http://hentaiathome.net/");
				}
			} else if(setting.equals("cur_client_build")) {
				if(Integer.parseInt(value) > CLIENT_BUILD) {
					warnNewClient = true;
				}
			} else if(setting.equals("server_time")) {
				serverTimeDelta = Integer.parseInt(value) - (int) (System.currentTimeMillis() / 1000);
				Out.debug("Setting altered: serverTimeDelta=" + serverTimeDelta);
				return true;
			}
			else if(setting.equals("rpc_server_ip")) {
				synchronized(rpcChangeMonitor) {
					String[] split = value.split(";");
					rpcServers = new java.net.InetAddress[split.length];
					int i = 0;
					for(String s : split) {
						rpcServers[i++] = java.net.InetAddress.getByName(s);
					}
				}
			}
			else if(setting.equals("image_server")) {
				imageServer = value;
			}
			else if(setting.equals("name")) {
				clientName = value;
			}
			else if(setting.equals("host")) {
				clientHost = value;
			}
			else if(setting.equals("port")) {
				clientPort = Integer.parseInt(value);
			}
			else if(setting.equals("request_server")) {
				requestServer = value;
			}
			else if(setting.equals("request_proxy_mode")) {
				requestProxyMode = Integer.parseInt(value);
			}
			else if(setting.equals("throttle_bytes")) {
				// THIS SHOULD NOT BE ALTERED BY THE CLIENT AFTER STARTUP. Using the website interface will update the throttle value for the dispatcher first, and update the client on the first stillAlive test.
				throttle_bytes = Integer.parseInt(value);
			}
			else if(setting.equals("hourbwlimit_bytes")) {
				// see above
				hourbwlimit_bytes = Long.parseLong(value);
			}
			else if(setting.equals("disklimit_bytes")) {
				long newLimit = Long.parseLong(value);

				if(newLimit >= disklimit_bytes) {
					disklimit_bytes = newLimit;
				}
				else {
					Out.warning("The disk limit has been reduced. However, this change will not take effect until you restart your client.");
				}
			}
			else if(setting.equals("diskremaining_bytes")) {
				diskremaining_bytes = Long.parseLong(value);
			}
			else if(setting.equals("force_dirty")) {
				forceDirty = value.equals("true");
			}
			else if(setting.equals("verify_cache")) {
				verifyCache = value.equals("true");
			}
			else if(setting.equals("use_less_memory")) {
				useLessMemory = value.equals("true");
			}
			else if(setting.equals("disable_logging")) {
				Out.disableLogging();
			}
			else if(setting.equals("disable_bwm")) {
				disableBWM = value.equals("true");
			}
			else if(setting.equals("skip_free_space_check")) {
				skipFreeSpaceCheck = value.equals("true");
			}
			else if(setting.equals("max_connections")) {
				overrideConns = Integer.parseInt(value);
			}
			else if(setting.equals("static_ranges")) {
				staticRanges = new Hashtable<String,Integer>();
				String[] split = value.split(";");
				for(String s : split) {
					if(s.length() == 4) {
						staticRanges.put(s, 1);
					}
				}
			}
			else if(!setting.equals("silentstart")) {
				// don't flag errors if the setting is handled by the GUI
				Out.warning("Unknown setting " + setting + " = " + value);
				return false;
			}

			Out.debug("Setting altered: " + setting +"=" + value);
			return true;
		} catch(Exception e) {
			Out.warning("Failed parsing setting " + setting + " = " + value);
		}

		return false;
	}

	public static void initializeDataDir() throws java.io.IOException {
		datadir = FileTools.checkAndCreateDir(new File("data"));
	}

	// accessor methods
	public static File getDataDir() {
		return datadir;
	}

	public static int getClientID() {
		return clientID;
	}

	public static String getClientKey() {
		return clientKey;
	}

	public static String getImageServer(String fileid) {
		return imageServer;
	}

	public static String getClientName() {
		return clientName;
	}

	public static String getClientHost() {
		return clientHost;
	}

	public static int getClientPort() {
		return clientPort;
	}

	public static String getRequestServer() {
		return requestServer;
	}

	public static int getRequestProxyMode() {
		return requestProxyMode;
	}

	public static int getThrottleBytesPerSec() {
		return throttle_bytes;
	}

	public static long getHourBWLimitBytes() {
		return hourbwlimit_bytes;
	}

	public static long getDiskLimitBytes() {
		return disklimit_bytes;
	}

	public static long getDiskMinRemainingBytes() {
		return diskremaining_bytes;
	}

	public static int getServerTime() {
		return (int) (System.currentTimeMillis() / 1000) + serverTimeDelta;
	}

	public static String getOutputLogPath() {
		return "data/log_out";
	}

	public static String getErrorLogPath() {
		return "data/log_err";
	}

	public static boolean isForceDirty() {
		return forceDirty;
	}

	public static boolean isVerifyCache() {
		return verifyCache;
	}

	public static boolean isUseLessMemory() {
		return useLessMemory;
	}

	public static boolean isSkipFreeSpaceCheck() {
		return skipFreeSpaceCheck;
	}

	public static boolean isWarnNewClient() {
		return warnNewClient;
	}

	public static boolean isDisableBWM() {
		return disableBWM;
	}

	public static HentaiAtHomeClient getActiveClient() {
		return activeClient;
	}

	public static HathGUI getActiveGUI() {
		return activeGUI;
	}

	public static boolean isValidRPCServer(InetAddress compareTo) {
		synchronized(rpcChangeMonitor) {
			if(rpcServers == null) {
				return false;
			}

			for(InetAddress i : rpcServers) {
				if(i.equals(compareTo)) {
					return true;
				}
			}

			return false;
		}
	}

	public static int getMaxConnections() {
		if(overrideConns > 0) {
			return overrideConns;
		}
		else {
			int conns = 0;
			int uptime = Stats.getUptime();

			if(throttle_bytes > 0) {
				conns = MAX_CONNECTION_BASE + (int) (throttle_bytes / 10000);
			} else if(uptime > 0) {
				// to be safe, we'll assume that each connection takes 120 seconds to finish. so 1 connection per second = 120 connections.
				conns = (int) (Stats.getFilesSent() * 120 / uptime);
			}

			return Math.max(Math.min(500, conns), MAX_CONNECTION_BASE);
		}
	}

	public static boolean isStaticRange(String fileid) {
		if(staticRanges != null) {
			return staticRanges.containsKey(fileid.substring(0, 4));
		}

		return false;
	}

	public static int getStaticRangeCount() {
		if(staticRanges != null) {
			return staticRanges.size();
		}

		return 0;
	}
}
*/
