package api

import (
	"../cache"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
)

var keyRegexp = regexp.MustCompile("^[0-9]{6}-[a-z0-9]{40}$")

// these functions do not communicate with the RPC server, but are actions triggered by it through servercmd
func (api *API) downloadFilesFromServer(files map[string]string) (returnText string, err error) {
	for file, key := range files {
		s := strings.Split(file, ":")
		fileid := s[0]
		host := s[1]

		// verify that we have valid ID and Key before we build an URL from it, in case the server has been compromised somehow...
		if cache.IsValidId(fileid) && keyRegexp.MatchString(key) {
			source := &url.URL{
				Scheme:   "http",
				Host:     fmt.Sprintf("%s:%d", host, 80),
				Path:     "image.php",
				RawQuery: "f=" + fileid + "&t=" + key,
			}

			if api.downloadAndCacheFile(source, fileid) {
				returnText += fileid + ":OK\n"
			} else {
				returnText += fileid + ":FAIL\n"
			}
		} else {
			returnText += fileid + ":INVALID\n"
		}
	}
	//} catch(Exception e) {
	//e.printStackTrace();
	//log.warning("Encountered error " + e + " when downloading image files from server. Will not retry.");
	//}

	return
}
func (api *API) doThreadedProxyTest(ipaddr string, port, count int, testsize, testtime int64, testkey int) string {
	successfulTests := 0
	totalTimeMillis := int64(0)

	log.Debugf("Running threaded proxy test against ipaddr=%s port=%d testsize=%d testcount=%d testtime=%d testkey=%s",
		ipaddr, port, testsize, count, testtime, testkey)

	for i := 0; i < count; i++ {
		source := &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", ipaddr, port),
			Path:   fmt.Sprintf("/t/%d/%d/%d/%d", testsize, testtime, testkey, rand.Int()),
		}
		log.Debugf("Test thread: %s", source)
	}

	/*
		try {
			List<FileDownloader> testfiles = Collections.checkedList(new ArrayList<FileDownloader>(), FileDownloader.class);

			for(int i=0; i<testcount; i++) {
				URL source = new URL("http", ipaddr, port, "/t/" + testsize + "/" + testtime + "/" + testkey + "/" + (int) Math.floor(Math.random() * Integer.MAX_VALUE));
				log.debug("Test thread: " + source);
				FileDownloader dler = new FileDownloader(source, 10000, 60000);
				testfiles.add(dler);
				dler.startAsyncDownload();
			}

			for(FileDownloader dler : testfiles) {
				if(dler.waitAsyncDownload()) {
					successfulTests += 1;
					totalTimeMillis += dler.getDownloadTimeMillis();
				}
			}
		} catch(java.net.MalformedURLException e) {
			HentaiAtHomeClient.dieWithError(e);
		}
	*/

	return fmt.Sprintf("OK:%d-%d", successfulTests, totalTimeMillis)
}

func (api *API) doProxyTest(ipaddr string, port int, fileid, keystamp string) string {
	/*
		if(!HVFile.isValidHVFileid(fileid)) {
			log.error("Encountered an invalid fileid in doProxyTest: " + fileid);
			return fileid + ":INVALID-0";
		}

		try {
			URL source = new URL("http", ipaddr, port, "/h/" + fileid + "/keystamp=" + keystamp + "/test.jpg");
			log.info("Running a proxy test against " + source + ".");

			// determine the approximate ping time to the other client
			// (if available, done on a best-effort basis).
			// why isn't there a built-in ping in java anyway?
			int pingtime = 0;

			// juuuuuust in case someone manages to inject a faulty IP address, we don't want to pass that unsanitized to an exec
			if(!ipaddr.matches("^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}$")) {
				log.warning("Invalid IP address: " + ipaddr);
			}
			else {
				// make an educated guess on OS to access the built-in ping utility
				String pingcmd = null;
				String whichOS = System.getProperty("os.name");

				if(whichOS != null) {
					if(whichOS.toLowerCase().indexOf("windows") > -1) {
						// windows style
						pingcmd = "ping -n 3 " + ipaddr;
					}
				}

				if(pingcmd == null) {
					// linux/unix/bsd/macos style
					pingcmd = "ping -c 3 " + ipaddr;
				}

				Process p = null;
				InputStreamReader isr = null;
				BufferedReader br = null;
				int pingresult = 0;
				int pingcount = 0;

				try {
					p = java.lang.Runtime.getRuntime().exec(pingcmd);
					isr = new InputStreamReader(p.getInputStream());
					br = new BufferedReader(isr);

					String read = null;

					while((read = br.readLine()) != null) {
						// try to parse the ping result and extract the result.
						// this will work as long as the time is enclosed between "time=" and "ms",
						// which it should be both in windows and linux. YMMV.
						int indexTime = read.indexOf("time=");

						if(indexTime >= 0) {
							int indexNumStart = indexTime + 5;
							int indexNumEnd = read.indexOf("ms", indexNumStart);

							if(indexNumStart > 0 && indexNumEnd > 0) {
								// parsing as double then casting, since linux gives a decimal number while windows doesn't
								pingresult += (int) Double.parseDouble(read.substring(indexNumStart, indexNumEnd).trim());
								++pingcount;
							}
						}
					}

					if(pingcount > 0) {
						pingtime = pingresult / pingcount;
					}
				} catch(Exception e) {
					log.debug("Encountered exception " + e + " while trying to ping remote client");
				} finally {
					try { br.close(); isr.close(); p.destroy(); } catch(Exception e) {}
				}
			}

			if(pingtime > 0) {
				log.debug("Approximate latency determined as ~" + pingtime + " ms");
			}
			else {
				log.debug("Could not determine latency, conservatively guessing 20ms");
				pingtime = 20;	// little to no compensation
			}

			long startTime = System.currentTimeMillis();

			if(downloadAndCacheFile(source, fileid)) {
				// this is mostly trial-and-error. we cut off 3 times the ping directly for TCP overhead (TCP three-way handshake + request/1st byte delay) , as well as cut off a factor of (1 second - pingtime) . this is capped to 200ms ping.
				long dlMillis = System.currentTimeMillis() - startTime;
				pingtime = Math.min(200, pingtime);
				double dlTime = Math.max(0, ((dlMillis * (1.0 - pingtime / 1000.0) - pingtime * 3) / 1000.0));
				log.debug("Clocked a download time of " + dlMillis + " ms. Ping delay fiddling reduced estimate to " + dlTime + " seconds.");
				return fileid + ":OK-" + dlTime;
			}
		} catch(Exception e) {
			log.warning("Encountered error " + e + " when doing proxy test against " + ipaddr + ":" + port + " on file " + fileid + ". Will not retry.");
		}

		return fileid + ":FAIL-0";
	*/

	return ""
}

// used by doProxyTest and downloadFilesFromServer
func (api *API) downloadAndCacheFile(source *url.URL, fileid string) bool {
	if cache.IsValidId(fileid) {
		/* TODO
		CacheHandler ch = client.getCacheHandler();
		File tmpfile = new File(ch.getTmpDir(), fileid);

		if(tmpfile.exists()) {
			tmpfile.delete();
		}

		FileDownloader dler = new FileDownloader(source, 10000, 30000);

		if(dler.saveFile(tmpfile)) {
			HVFile hvFile = HVFile.getHVFileFromFile(tmpfile, true);

			if(hvFile != null) {
				if(!hvFile.getLocalFileRef().exists()) {
					if(ch.moveFileToCacheDir(tmpfile, hvFile)) {
						ch.addFileToActiveCache(hvFile);
						log.info("The file " + fileid + " was successfully downloaded and inserted into the active cache.");
					}
					else {
						log.warning("Failed to insert " + fileid + " into cache.");
						tmpfile.delete();
						// failed to move, but didn't exist.. so we'll fail
						return false;
					}
				}
				else {
					log.info("The file " + fileid + " was successfully downloaded, but already exists in the cache.");
					tmpfile.delete();
				}

				// if the file was inserted, or if it exists, we'll call it a success
				Stats.fileRcvd();
				return true;
			}
			else {
				log.warning("Downloaded file " + fileid + " failed hash verification. Will not retry.");
			}
		}
		else {
			log.warning("Failed downloading file " + fileid + " from " + source + ". Will not retry.");
		}

		if(tmpfile.exists()) {
			tmpfile.delete();
		}
		*/
	} else {
		log.Warnf("Encountered invalid fileid %s", fileid)
	}

	return false
}
