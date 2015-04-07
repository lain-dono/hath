package main

import (
	"./api"
	"./cache"
	"./util"
	"github.com/Sirupsen/logrus"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var Client *api.API
var CacheHandler *cache.Handler

var log = util.Logger()

func main() {
	done := exitHandler()

	defer func() {
		err := recover()
		if err != nil {
			//log.Fatal("Failed to initialize InputQueryHandler")
		}
	}()
	log.Infof(copyright, api.CLIENT_VERSION, api.CLIENT_BUILD)

	log.Info("Logging in to main server...")

	log.Info("Initializing cache handler...")
	CacheHandler = cache.NewHandler("./cache", "./tmp", "data.db")
	log.Info(CacheHandler)

	Client = api.New(api.CLIENT_API_URL)

	if login, err := loadClientLoginFromFile("client_login"); err == nil {
		log.Infof("Loaded login settings from %s", "client_login")
		Client.Login = login
	}
	//id, key := promptForIDAndKey()
	//log.Print(id, key)

	//if !Settings.loginCredentialsAreSyntaxValid() {
	//Settings.promptForIDAndKey(iqh);
	//}

	log.Infof("Connecting to the Hentai@Home Server to register client with ID %d...", Client.Id)
	s, err := Client.RefreshServerStat()
	if err != nil {
		log.Fatal("Failed to get initial stat from server. ", err)
	}
	parseAndUpdateSettings(s)

	for {
		settings, _ := Client.LoadClientSettingsFromServer()
		parseAndUpdateSettings(settings)
		break
	}

	//CacheHandler.Initialize()
	//CacheHandler.FlushRecentlyAccessed()

	//

	//log.Print(s)

	//InputQueryHandler iqh = nil;

	//try {
	//iqh = InputQueryHandlerCLI.getIQHCLI();
	//new HentaiAtHomeClient(iqh, args);
	//} catch(Exception e) {
	//Out.error("Failed to initialize InputQueryHandler");
	//}

	for i := 0; i < 5; i++ {
		//log.Debug("vfds")
		//log.Info("vfdsnjkl")
		//log.Warn("-")
		//log.Error("-")
	}

	log.Info("WAIT")
	<-done
	log.Print("done")

	rand.Seed(time.Now().UnixNano())
	if rand.Float64() > 0.99 {
		log.Info(heart)
	} else {
		log.Info(exitMsgs[rand.Intn(len(exitMsgs))])
	}

	return

	// FATAL

	for {
		log.WithFields(logrus.Fields{
			"animal": "walrus",
			"size":   "10",
		}).Print("To invoke the hive-mind representing chaos.")

		log.WithFields(logrus.Fields{
			"omg":    true,
			"number": 122,
		}).Warn("Invoking the feeling of chaos.")

		log.WithFields(logrus.Fields{
			"animal": "walrus",
			"size":   "10",
		}).Print("With out order.")

		log.WithFields(logrus.Fields{
			"animal": "walrus",
			"size":   "9",
		}).Error("The Nezperdian hive-mind of chaos. Zalgo.")

		log.WithFields(logrus.Fields{
			"omg":    true,
			"number": 100,
		}).Warn("He who Waits Behind The Wall.")

		log.Fatal("ZALGO !")
	}
}

func exitHandler() (done chan bool) {
	done = make(chan bool, 1)

	sigs := make(chan os.Signal, 1)
	//signal.Notify(sigs, syscall.Interrupt, syscall.SIGTERM)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigs {
			log.Warnf("captured %v", sig)
			//os.Exit(1)
			done <- true
		}
	}()
	return
}

const heart = `
                             .,---.
                           ,/XM#MMMX;,
                         -%##########M%,
                        -@######%  $###@=
         .,--,         -H#######$   $###M:
      ,;$M###MMX;     .;##########$;HM###X=
    ,/@##########H=      ;################+
   -+#############M/,      %##############+
   %M###############=      /##############:
   H################      .M#############;.
   @###############M      ,@###########M:.
   X################,      -$=X#######@:
   /@##################%-     +######$-
   .;##################X     .X#####+,
    .;H################/     -X####+.
      ,;X##############,       .MM/
         ,:+$H@M#######M#$-    .$$=
              .,-=;+$@###X:    ;/=.
                     .,/X$;   .::,
                         .,    ..
`

var exitMsgs = []string{
	"I don't hate you",
	"Whyyyyyyyy...",
	"No hard feelings",
	"Your business is appreciated",
	"Good-night",
}

const copyright = `Hentai@Home %s:%d starting up

	Copyright (c) 2008-2014, E-Hentai.org - all rights reserved.
	This software comes with ABSOLUTELY NO WARRANTY.
	This is free software, and you are welcome to modify and redistribute it under the GPL v3 license.
`
