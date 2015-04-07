// +build ignore

package cache

import (
	log "../log"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"time"
	//_ "github.com/mattn/go-sqlite3"
	//"expvar"
	"sync"
)

const (
	MEMORY_TABLE_ELEMENTS = 1048576
)

//var Stats = expvar.NewMap("Stats")
//var cacheCount = expvar.NewInt("Stats.cacheCount")
//var cacheSize = expvar.NewInt("Stats.cacheSize")

//var _ = expvar.NewVar("CacheHandler", o)
//var handler = &Handler{}

//func init() {
//expvar.Publish("CacheHandler", handler)
//}

type HentaiAtHomeClient interface {
	DieWithError(err error)
}

//package org.hath.base;

//import java.io.File;
//import java.util.ArrayList;
//import java.util.Collections;
//import java.util.LinkedList;
//import java.util.List;
//import java.sql.*;
type Handler struct {
	cachedirPath, tmpdirPath string
	cachedir, tmpdir         *os.File

	client HentaiAtHomeClient

	cacheCount, startupCachedFileStrlen int

	cacheSize  int64
	quickStart bool

	pendingRegister       []Id
	recentlyAccessed      []CachedFile
	recentlyAccessedFlush int64

	memoryWrittenTable []uint16
	memoryClearPointer int

	db *DB

	sync.Mutex
}

func (h *Handler) String() string { return "" }

func NewHandler() (handler *Handler, err error) {
	//tmpdir = FileTools.checkAndCreateDir(new File("tmp"));
	//cachedir = FileTools.checkAndCreateDir(new File("cache"));

	handler = &Handler{
		//client:           client,
		recentlyAccessed: make([]CachedFile, 0, 100),
		pendingRegister:  make([]Id, 0, 50),
	}

	if !Settings.isUseLessMemory() {
		// the memoryWrittenTable can hold 16^5 = 1048576 shorts consisting of 16 bits each.
		// addressing is done by looking up the first five nibbles (=20 bits) of a hash,
		//   then using the sixth nibble to determine which bit in the short to read/set.
		// while collisions may occur, they should be fairly rare,
		//   and should not cause any major issues with files not having their timestamp updated.
		// (and even if it does, the impact of this will be negligible,
		//   as it will only cause the LRU mechanism to be slightly less efficient.)
		handler.memoryWrittenTable = make([]uint16, MEMORY_TABLE_ELEMENTS)
		handler.memoryClearPointer = 0
	} else {
		handler.memoryWrittenTable = nil
	}

	log.Info("CacheHandler: Initializing database engine...")

	// TODO defer
	//if err = initializeDatabase("data/hath.db"); err != nil {
	//log.Error("CacheHandler: Failed to initialize SQLite database engine")
	//client.DieWithError(err)
	//}

	if err = handler.initializeDatabase("data/hath.db"); err != nil {
		log.Info("")
		log.Info("**************************************************************************************************************")
		log.Info("The database could not be loaded. Please check file permissions and file system integrity.")
		log.Info("If everything appears to be working, please do the following:")
		log.Info("1. Locate the directory " + Settings.DataDirAhsolutePath())
		log.Info("2. Delete the file hath.db")
		log.Info("3. Restart the client.")
		log.Info("The system should now rebuild the database.")
		log.Info("***************************************************************************************************************")
		log.Info("")

		//client.DieWithError(errors.New("Failed to load the database."))
		err = errors.New("Failed to load the database.")
		return
	}

	if handler.quickStart {
		log.Info("Last shutdown was clean - using fast startup procedure.")
	} else {
		log.Info("Last shutdown was dirty - the cache index must be verified.")
	}

	log.Info("CacheHandler: Database initialized")

	return
}

func (h *Handler) initializeDatabase(db string) (err error) {
	defer func() {
		if err != nil {
			log.Error("CacheHandler: Encountered error reading database.")
			h.db.Terminate()
			// print stack
			panic(err)
		}
	}()

	h.db = NewDB(db)
	/*
		if err != nil {
			return
		}
	*/

	h.resetFutureLasthits()

	h.db.Optimize()

	// are we dirty?
	if !Settings.isForceDirty() {
		val, e := h.db.CleanShutdown()
		if e != nil {
			return e
		}
		h.quickStart = val == CLEAN_SHUTDOWN_VALUE
	}
	err = h.db.SetCleanShutdown(currentTimeMs())

	return
}

func (h *Handler) resetFutureLasthits() (err error) {
	//nowtime := currentTimeS()

	//sqlite.setAutoCommit(false);

	log.Info("CacheHandler: Checking future lasthits on non-static files...")
	/*
		PreparedStatement getFutureLasthits = sqlite.prepareStatement("SELECT fileid FROM CacheList WHERE lasthit>?;");
		getFutureLasthits.setLong(1, nowtime + 2592000);
		ResultSet rs = getFutureLasthits.executeQuery();

		List<String> removelist = Collections.checkedList(new ArrayList<String>(), String.class);

		while( rs.next() ) {
			String fileid = rs.getString(1);
			if( !Settings.isStaticRange(fileid) ) {
				removelist.add(fileid);
			}
		}

		rs.close();

		//PreparedStatement resetLasthit = sqlite.prepareStatement("UPDATE CacheList SET lasthit=? WHERE fileid=?;");

		for( String fileid : removelist ) {
			//resetLasthit.setLong(1, nowtime);
			//resetLasthit.setString(2, fileid);
			//resetLasthit.executeUpdate();

			deleteCachedFile.setString(1, fileid);
			deleteCachedFile.executeUpdate();
			HVFile.getHVFileFromId(fileid).getLocalFileRef().delete();
			Out.debug("Removed old static range file " + fileid);
		}

		log.Info("CacheHandler: Resetting remaining far-future lasthits...");

		PreparedStatement resetFutureStatic = sqlite.prepareStatement("UPDATE CacheList SET lasthit=? WHERE lasthit>?;");

		resetFutureStatic.setLong(1, nowtime + 7776000);
		resetFutureStatic.setLong(2, nowtime + 31536000);
		resetFutureStatic.executeUpdate();

		sqlite.setAutoCommit(true);
	*/
	return
}

func (h *Handler) initializeCacheHandler() (err error) {
	log.Info("CacheHandler: Initializing the cache system...")

	// delete orphans from the temp dir

	tmpfiles, err := h.tmpdir.Readdir(0)
	if err != nil {
		return
	}

	for _, file := range tmpfiles {
		if file.Mode().IsRegular() {
			log.Debugf("Deleted orphaned temporary file %s", file.Name())
			os.Remove(path.Join(h.tmpdirPath, file.Name()))
		} else {
			log.Warn("Found a non-file %s in the temp directory, won't delete.", file.Name())
		}
	}

	if h.quickStart && !Settings.isVerifyCache() {
		c, s, err := h.db.CountStats()
		if err != nil {
			log.Error("CacheHandler: Failed to perform database operation")
			h.client.DieWithError(err)
		}
		h.cacheCount, h.cacheSize = c, s

		h.updateStats()
		h.flushRecentlyAccessed()
	} else {
		if Settings.isVerifyCache() {
			log.Info("CacheHandler: A full cache verification has been requested. This can take quite some time.")
		}

		h.populateInternalCacheTable()
	}

	/* TODO
		if !Settings.isSkipFreeSpaceCheck() && (cachedir.getFreeSpace() < Settings.getDiskLimitBytes()-cacheSize) {
			// NOTE: if this check is removed and the client ends up
			// being starved on disk space with static ranges assigned,
			// it will cause a major loss of trust.
			h.client.setFastShutdown()
			h.client.dieWithError(`The storage device does not have enough space available to hold the given cache size.
	Free up space, or reduce the cache size from the H@H settings page.
	http://g.e-hentai.org/hentaiathome.php?cid=` + Settings.ClientID())
		}

		if (h.cacheCount < 1) && (Settings.getStaticRangeCount() > 0) {
			// NOTE: if this check is removed and the client is started
			// with an empty cache and several static ranges assigned,
			// it will cause a major loss of trust.
			h.client.setFastShutdown()
			h.client.DieWithError(`This client has static ranges assigned to it, but the cache is empty.
	Check permissions and, if necessary, delete the file hath.db in the data directory to rebuild the cache database.
	If the cache has been deleted or is otherwise lost, you have to manually reset your static ranges from the H@H settings page.
	http://g.e-hentai.org/hentaiathome.php?cid=` + Settings.ClientID())
		}
	*/

	// TODO
	//if !h._checkAndFreeDiskSpace(h.cachedir, true) {
	//log.Warn("ClientHandler: There is not enough space left on the disk to add more files to the cache.")
	//}

	return
}

/*
func (h *Handler) HVFile(fileid string, hit bool) (hvf HVFile, ok bool) {
	ok = IsValidId(fileid)
	//if !IsValidId(fileid) {
	if ok {
		cf := NewCachedFile(fileid)

		if hit {
			cf.Hit()
		}

		hvf, _ = cf.HVFile()
	}
	return
}
*/

// NOTE: this will just move the file into its correct location.
// addFileToActiveCache MUST be called afterwards to import the file into the necessary datastructures.
// otherwise, the file will not be available until the client is restarted, and even then not if --quickstart is used.
func (h *Handler) moveFileToCacheDir( /*file *os.File*/ from string, id Id) (err error) {
	to := h.ResolveId(id)
	// TODO
	//if h.checkAndFreeDiskSpace(file) {
	return os.Rename(from, to)

	/*
		toFile, err := hvFile.LocalFileRef()
		if err != nil {
			//e.printStackTrace();
			log.Warn("CacheHandler: Encountered exception %s when moving file %s", err, file.Name())
			panic(err)
			//}
			return false
		}

		//try {
		// TODO FileTools.checkAndCreateDir(toFile.ParentFile())

		switch {
		case file.renameTo(toFile):
			log.Debug("CacheHandler: Imported file %s to %s", file, hvFile.Id())
			return true
		case FileTools.copy(file, toFile):
			// rename can fail in some cases, like when source and target are on different file systems.
			// when this happens, we just use our own copy function instead, and delete the old file afterwards.
			//removeFile
			file.delete() // TODO
			log.Debug("CacheHandler: Imported file %s to %s", file.Name(), hvFile.Id())
			return true
		default:
			log.Warnf("CacheHandler: Failed to move file %s", file.Name())
		}
	*/
	//}
	return fmt.Errorf("fail move (disk full) %s to %s", from, to)
}
func (h *Handler) addFileToActiveCache(id Id) (err error) {
	h.Lock()
	defer h.Unlock()

	//try {
	//synchronized(sqlite) {
	fileid := id.String()
	affected, err := h.db.UpdateCachedFileActive(fileid)
	if affected == 0 {
		lasthit := time.Now()

		if Settings.isStaticRange(fileid) {
			// if the file is in a static range, bump to three months in the future.
			// on the next access, it will get bumped further to a year.
			lasthit = lasthit.Add(time.Hour * 24 * 90)
		}

		h.db.InsertCachedFile(id, lasthit)
	}

	h.cacheCount++
	h.cacheSize += id.Size()
	h.updateStats()

	return
}

// During server-initiated file distributes and proxy tests against other clients,
// the file is automatically registered for this client by the server,
// but this doesn't happen during client-initiated H@H Downloader or H@H Proxy downloads.
// So we'll instead send regular updates to the server about downloaded files, whenever a file is added this way.
func (h *Handler) addPendingRegisterFile(id Id) (err error) {
	h.Lock()
	defer h.Unlock()
	// We only register files <= 10 MB. Larger files are handled outside the H@H network.

	if (id.Size() <= 10485760) && !h.db.IsStaticRange(id.Hash()) {
		log.Debugf("Added %s to pendingRegister", id)
		h.pendingRegister = append(h.pendingRegister, id)

		if len(h.pendingRegister) >= 50 {
			// this call also empties the list
			// TODO h.client.ServerHandler().notifyRegisterFiles(h.pendingRegister)
		}
	} else {
		log.Debugf("Not registering file %s - in static range or larger than 10 MB", id)
	}
	return
}
func (h *Handler) deleteFileFromCache(toRemove Id) (err error) {
	h.Lock()
	defer h.Unlock()
	return h.deleteFileFromCacheNosync(toRemove)
}

func (h *Handler) deleteFileFromCacheNosync(toRemove Id) (err error) {
	//try {
	err = h.db.DeleteCachedFile(toRemove)
	if err != nil {
		return
	}
	h.cacheCount--
	h.cacheSize -= toRemove.Size()

	// TODO
	err = removeFile(h.ResolveId(toRemove))
	log.Info("CacheHandler: Deleted cached file %s", toRemove)
	h.updateStats()
	return
}

type byName []os.FileInfo

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }

func (h *Handler) populateInternalCacheTable() (err error) {

	//try {
	h.db.cacheIndexClearActive.Exec()
	h.cacheCount = 0
	h.cacheSize = 0

	knownFiles := 0
	newFiles := 0

	// load all the files directly from the cache directory itself and
	// initialize the stored last access times for each file.
	// last access times are used for the LRU-style cache.

	log.Info("CacheHandler: Loading cache.. (this could take a while)")

	scdirs, err := h.cachedir.Readdir(0)
	if err != nil {
		return
	}
	sort.Sort(byName(scdirs))

	//try {

	// we're doing some SQLite operations here without synchronizing on the SQLite connection. the program is single-threaded at this point, so it should not be a real problem.
	loadedFiles := 0
	//sqlite.setAutoCommit(false);

	for _, scdir := range scdirs {
		if scdir.IsDir() {
			cdirs, e := h.cachedir.Readdir(0)
			if e != nil {
				return e
			}
			sort.Sort(byName(cdirs))

			for _, cfile := range cdirs {
				newFile := false
				/* TODO
				synchronized(sqlite) {
					queryCachedFileLasthit.setString(1, cfile.getName());
					ResultSet rs = queryCachedFileLasthit.executeQuery();
					newFile = !rs.next();
					rs.close();
				}
				*/

				hvFile, err := NewHVFileFromId(cfile.Name()) // TODO , Settings.isVerifyCache() || newFile)
				// TODO verify
				if err != nil {
					// TODO
				}

				if hvFile != nil {
					h.addFileToActiveCache(hvFile)

					if newFile {
						newFiles++
						log.Infof("CacheHandler: Verified and loaded file %s", cfile.Name())
					} else {
						knownFiles++
					}

					loadedFiles++
					if loadedFiles%1000 == 0 {
						log.Infof("CacheHandler: Loaded %d files so far...", loadedFiles)
					}
				} else {
					log.Warn("CacheHandler: The file %s was corrupt. It is now deleted.", cfile.Name())
					removeFile(h.cachedirPath, scdir.Name(), cfile.Name())
				}
			}
		} else {
			removeFile(h.cachedirPath, scdir.Name())
		}

		h._flushRecentlyAccessed(false)
	}

	//sqlite.commit();
	//sqlite.setAutoCommit(true);

	// TODO
	//synchronized(sqlite) {
	//int purged = deleteCachedFileInactive.executeUpdate();
	//Out.info("CacheHandler: Purged " + purged + " nonexisting files from database.");
	//}
	//} catch(Exception e) {
	//Out.error("CacheHandler: Failed to perform database operation");
	//client.dieWithError(e);
	//}

	log.Infof("CacheHandler: Loaded %d known files.", knownFiles)
	log.Infof("CacheHandler: Loaded %d new files.", newFiles)
	log.Infof("CacheHandler: Finished initializing the cache (%d files, %d bytes)", h.cacheCount, h.cacheSize)

	h.updateStats()
	//} catch(Exception e) {
	//e.printStackTrace();
	//HentaiAtHomeClient.dieWithError("Failed to initialize the cache.");
	//}
	return
}

/*
func (h *Handler) recheckFreeDiskSpace() (err error) {
	return h._checkAndFreeDiskSpace(h.cachedir, false)
}

func (h *Handler) checkAndFreeDiskSpace(file *os.File) (err error) {
	return h._checkAndFreeDiskSpace(file, false)
}
*/

func (h *Handler) _checkAndFreeDiskSpace(file *os.File, noServerDeleteNotify bool) (deleteNotify []Id, err error) {
	if file == nil {
		h.client.DieWithError(errors.New("CacheHandler: checkAndFreeDiskSpace needs a file handle to calculate free space"))
		///
		return
	}

	bytesNeeded := int64(0)
	s, _ := file.Stat()
	if s.Mode().IsRegular() {
		bytesNeeded = s.Size()
	}
	cacheLimit := int64(Settings.getDiskLimitBytes())

	log.Debugf("CacheHandler: Checking disk space (adding %d bytes: cacheSize=%d, cacheLimit=%d, cacheFree=%d)",
		bytesNeeded, h.cacheSize, cacheLimit, (cacheLimit - h.cacheSize))

	// we'll free ten times the size of the file or 20 files, whichever is largest.
	bytesToFree := int64(0)

	if h.cacheSize > cacheLimit {
		bytesToFree = h.cacheSize - cacheLimit
	} else if h.cacheSize+bytesNeeded-cacheLimit > 0 {
		bytesToFree = bytesNeeded * 10
	}

	if bytesToFree > 0 {
		log.Infof("CacheHandler: Freeing at least %d bytes...", bytesToFree)
		//List<HVFile> deleteNotify = Collections.checkedList(new ArrayList<HVFile>(), HVFile.class);
		//var deleteNotify []*HVFile

		//try {
		for bytesToFree > 0 && h.cacheCount > 0 {
			//synchronized(sqlite) {
			ids, _, err := h.db.C.CCachedFileSortOnLasthit(0, 20)
			if err != nil {
				//
			}
			for _, fileid := range ids {
				toRemove, ok := NewIdFromString(fileid)

				if ok {
					h.deleteFileFromCacheNosync(toRemove)
					bytesToFree -= toRemove.Size()

					if !Settings.isStaticRange(fileid) {
						// don't notify for static range files
						deleteNotify = append(deleteNotify, toRemove)
					}
				}
			}

			//}
		}
		//}
		//catch(Exception e) {
		//Out.error("CacheHandler: Failed to perform database operation");
		//client.dieWithError(e);
		//}

		if !noServerDeleteNotify {
			//client.getServerHandler().notifyUncachedFiles(deleteNotify)
		}
		//*/
	}
	if Settings.isSkipFreeSpaceCheck() {
		log.Debug("CacheHandler: Disk free space check is disabled.")
		return
	} else {
		diskFreeSpace := getFreeSpace(file)

		if diskFreeSpace < max(Settings.getDiskMinRemainingBytes(), 104857600) {
			// if the disk fills up, we  stop adding files instead of starting to remove files from the cache,
			// to avoid being unintentionally squeezed out by other programs
			log.Warnf("CacheHandler: Cannot meet space constraints: Disk free space limit reached (%d bytes free on device)", diskFreeSpace)
			err = errors.New("NOOOOOOOOOOOOOOOOO")
			return
		} else {
			log.Debugf("CacheHandler: Disk space constraints met (%d bytes free on device)", diskFreeSpace)
			return
		}
	}
	return
}

func (h *Handler) pruneOldFiles() (deleteNotify []Id, err error) {
	h.Lock()
	defer h.Unlock()

	//List<HVFile> deleteNotify = Collections.checkedList(new ArrayList<HVFile>(), HVFile.class);
	pruneCount := 0

	log.Info("Checking for old files to prune...")

	ids, hits, err := h.db.CachedFileSortOnLasthit(0, 20)
	if err != nil {
		//
	}
	nowtime := time.Now()
	for i, fileid := range ids {
		lasthit := hits[i]

		if lasthit < nowtime-2592000 {
			toRemove, ok := NewIdFromString(fileid)

			if ok {
				h.deleteFileFromCacheNosync(toRemove)
				pruneCount++

				if !Settings.isStaticRange(fileid) {
					// don't notify for static range files
					deleteNotify = append(deleteNotify, toRemove)
				}
			}
		}
	}
	//}
	//catch(Exception e) {
	//Out.error("CacheHandler: Failed to perform database operation");
	//client.dieWithError(e);
	//}

	//client.getServerHandler().notifyUncachedFiles(deleteNotify)

	log.Infof("Pruned %d files.", pruneCount)
	return
}

//String[] blacklisted = client.getServerHandler().getBlacklist(deltatime);
func (h *Handler) processBlacklist(blacklisted []string, noServerDeleteNotify bool) (deleteNotify []Id, err error) {
	log.Info("CacheHandler: Retrieving list of blacklisted files...")

	if len(blacklisted) == 0 {
		log.Warn("CacheHandler: Failed to retrieve file blacklist, will try again later.")
		return
	}

	log.Info("CacheHandler: Looking for and deleting blacklisted files...")

	counter := 0
	//List<HVFile> deleteNotify = Collections.checkedList(new ArrayList<HVFile>(), HVFile.class);

	//try {
	//synchronized(sqlite) {
	for _, fileid := range blacklisted {
		_, err := h.db.CachedFileLasthit(fileid)
		//var toRemove HVFile = nil
		//_, err := queryCachedFileLasthit.setString(1, fileid)

		if err == nil {
			toRemove, _ := NewHVFileFromId(fileid)
			log.Infof("CacheHandler: Removing blacklisted file %s", fileid)
			counter++
			h.deleteFileFromCacheNosync(toRemove)

			if !Settings.isStaticRange(toRemove.Id()) {
				// do not notify about static range files
				deleteNotify = append(deleteNotify, toRemove)
			}
		}
	}
	//}
	//} catch(Exception e) {
	//Out.error("CacheHandler: Failed to perform database operation");
	//client.dieWithError(e);
	//}

	if !noServerDeleteNotify {
		//client.getServerHandler().notifyUncachedFiles(deleteNotify)
	}

	log.Info("CacheHandler: %d blacklisted files were removed.", counter)
	return
}

func (h *Handler) updateStats() {
	// TODO
	//cacheCount.Set(h.cacheCount)
	//cacheSize.Set(h.cacheSize)
}

func (h *Handler) CacheSize() int64 {
	return h.cacheSize
}
func (h *Handler) CacheCount() int {
	return h.cacheCount
}

func (h *Handler) SegmentCount() int {
	if (Settings.isUseLessMemory() && h.cacheCount > 16000) || (h.cacheCount > 400000) {
		return 256
	} else {
		return 16
	}
}

func (h *Handler) StartupCachedFilesStrlen() int {
	return h.startupCachedFileStrlen
}

func (h *Handler) calculateStartupCachedFilesStrlen() {
	segmentCount := h.SegmentCount()
	h.startupCachedFileStrlen = 0

	for segmentIndex := 0; segmentIndex < segmentCount; segmentIndex++ {
		//LinkedList<String> fileList = getCachedFilesSegment(Integer.toHexString(segmentCount | segmentIndex).substring(1));
		name := strconv.FormatInt(int64(segmentCount|segmentIndex), 16)
		fileList, _ := h.db.CachedFilesSegment(name[1:1]) // FIXME

		for _, fileid := range fileList {
			h.startupCachedFileStrlen += len(fileid) + 1
		}

		log.Infof("Calculated segment %d of %d", segmentIndex, segmentCount)
	}
}

func (h *Handler) CacheDir() string { //*os.File {
	return h.cachedirPath
}
func (h *Handler) TmpDir() string {
	return h.tmpdirPath
}

func (h *Handler) flushRecentlyAccessed() {
	h._flushRecentlyAccessed(true)
}
func (h *Handler) _flushRecentlyAccessed(disableAutocommit bool) {
	var flushCheck, flush []CachedFile

	if h.memoryWrittenTable != nil {
		// this function is called every 10 seconds.
		// clearing 121 of the shorts for each call means that each element
		// will live up to a day (since 1048576 / 8640 is roughly 121).
		// NOTE that this is skipped if the useLessMemory flag is set.

		clearUntil := min(MEMORY_TABLE_ELEMENTS, h.memoryClearPointer+121)

		//Out.debug("CacheHandler: Clearing memoryWrittenTable from " + memoryClearPointer + " to " + clearUntil);

		for h.memoryClearPointer < clearUntil {
			h.memoryClearPointer++
			h.memoryWrittenTable[h.memoryClearPointer] = 0
		}

		if clearUntil >= MEMORY_TABLE_ELEMENTS {
			h.memoryClearPointer = 0
		}
	}
	//synchronized(recentlyAccessed) {
	h.recentlyAccessedFlush = currentTimeMs()
	flushCheck = append(h.recentlyAccessed)
	h.recentlyAccessed = nil
	//}

	if len(flushCheck) == 0 {
		return
	}
	//try {
	//synchronized(sqlite) {
	//XXX flush = new ArrayList<CachedFile>(flushCheck.size());
	for _, cf := range flushCheck {
		fileid := cf.Id()
		doFlush := true

		if h.memoryWrittenTable != nil {
			// if the memory table is active, we use this as a first step
			// in order to determine if the timestamp should be updated or not.
			// we first need to compute the array index and bitmask for this particular fileid.
			// then, if the bit is set, we do not update. if not, we update but set the bit.

			doFlush = false

			//try {
			arrayIndex := 0
			for i := 0; i < 5; i++ {
				ii, _ := strconv.ParseUint(fileid[i:i+1], 16, 0)
				arrayIndex += int(ii << ((4 - uint64(i)) * 4))
			}

			n, _ := strconv.ParseUint(fileid[5:6], 16, 0)

			bitMask := uint16(1 << n)

			if (h.memoryWrittenTable[arrayIndex] & bitMask) != 0 {
				//Out.debug("Written bit for " + fileid + " = " + arrayIndex + ":" + fileid.charAt(5) + " was set");
			} else {
				//Out.debug("Written bit for " + fileid + " = " + arrayIndex + ":" + fileid.charAt(5) + " was not set - flushing");
				h.memoryWrittenTable[arrayIndex] |= bitMask
				doFlush = true
			}
			//} catch(Exception e) {
			//log.Warnf("Encountered invalid fileid %s while checking memoryWrittenTable.", fileid);
			//}
		}

		if doFlush {
			// we don't need higher resolution than a day for the LRU mechanism,
			// so we'll save expensive writes by not updating timestamps for files that have been flagged the previous 24 hours.
			// (reads typically don't involve an actual disk access as the database file
			// is cached to RAM - writes always do unless it can be combined with another write)

			hittime, err := h.db.CachedFileLasthit(fileid)
			if err != nil {
				err = nil
			} else {
				//long hittime = rs.getLong(1);
				nowtime := currentTimeS()

				if Settings.isStaticRange(fileid) {
					// for static range files, do not flush if it was already flushed this month
					// static range files have hittime set to nowtime + 31536000 every time they flush
					// so if hittime is less than nowtime + 28944000, flush it
					if hittime > nowtime+28944000 {
						doFlush = false
					}
				} else {
					// for other files, do not flush if it was already flushed today
					if hittime > nowtime-86400 {
						doFlush = false
					}
				}
			}

			if doFlush {
				flush = append(flush, cf)
			}
		}
	}
	if len(flush) > 0 {
		// FIXME
		//if(disableAutocommit) {
		//sqlite.setAutoCommit(false);
		//}

		for _, cf := range flush {
			if cf.NeedsFlush() {
				fileid := cf.Id()
				lasthit := currentTimeS()

				if Settings.isStaticRange(fileid) {
					// if the file is in a static range, bump to one year in the future
					lasthit += 31536000
				}
				h.db.SetCachedFileLasthit(lasthit, fileid)

				// there is a race condition here of sorts, but it doesn't matter.
				// flushed() will set needFlush to false, which can be set to true by hit(),
				// but no matter the end result we have an acceptable outcome.
				// (it's always flushed at least once.)
				cf.Flushed()
			}
		}

		//if(disableAutocommit) {
		//sqlite.setAutoCommit(true);
		//}
	}
	//} catch(Exception e) {
	//Out.error("CacheHandler: Failed to perform database operation");
	//client.dieWithError(e);
}
