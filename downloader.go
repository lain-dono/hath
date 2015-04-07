package main

import (
	"./cache"
	"./util"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func downloader(fileid, token string, gid, page int, fname string, skipHath bool) {
	id, ok := cache.NewIdFromString(fileid)
	if !ok {
		panic("fail id " + fileid)
	}

	source := &url.URL{
		Scheme: "http",
		Host:   "", // TODO
		Path:   fmt.Sprintf("/r/%s/%s/%d-%d/%s", fileid, token, gid, page, fname),
	}
	if skipHath {
		source.RawQuery = "nl=1"
	}
	log.Debug("GalleryFileDownloader: Requesting file download from ", source)

	conn := &http.Client{
		//Timeout: 30000,
		Timeout: 10000,
	}

	req := &http.Request{
		URL: source,
		Header: map[string][]string{
			"User-Agent":   {"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.12) Gecko/20080201 Firefox/2.0.0.12"},
			"Hath-Request": {strconv.FormatInt(Client.Id, 10) + "-" + util.SHA(Client.Key+fileid)},
		},
	}

	retry := true
	retval := 0

	for retry {
		retry = false

		resp, err := conn.Do(req)
		if err != nil {
			panic(err)
		}

		if resp.ContentLength < 0 {
			log.Warn("Request host did not send Content-Length, aborting transfer. (" + /*connection +*/ ")")
			log.Warn("Note: A common reason for this is running firewalls with outgoing restrictions or programs like PeerGuardian/PeerBlock. Verify that the remote host is not blocked.")
			//retval = 502;
		} else if resp.ContentLength > 10485760 {
			log.Warn("Content-Length is larger than 10 MB, aborting transfer. (" + /*connection +*/ ")")
			//retval = 502;
		} else if resp.ContentLength != id.Size() {
			//log.Warnf("Reported contentLength %d does not match expected length of file %s (%q)", contentLength, fileid, conn)

			// this could be more solid, but it's not important.
			// this will only be tested if there is a fail,
			// and even if the fail somehow matches the size of the error images,
			// the server won't actually increase the limit unless we're close to it.
			if retval == 0 && (resp.ContentLength == 28658 || resp.ContentLength == 1009) {
				log.Warn("We appear to have reached the image limit. Attempting to contact the server to ask for a limit increase...")
				// FIXME client.getServerHandler().notifyMoreFiles();
				retry = true
				retval = 502
			}
		} else {
			retval = 0
		}
	}
}
