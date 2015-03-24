package cache

import (
	log "../log"
	"database/sql"
)

type DB struct {
	//isForceDirty bool

	sqlite *sql.DB

	cacheIndexClearActive, cacheIndexCountStats                                 *sql.Stmt
	queryCachelistSegment, queryCachedFileLasthit, queryCachedFileSortOnLasthit *sql.Stmt
	insertCachedFile, updateCachedFileLasthit, updateCachedFileActive           *sql.Stmt
	deleteCachedFile, deleteCachedFileInactive                                  *sql.Stmt
	getStringVar, setStringVar                                                  *sql.Stmt
}

func (h *DB) check(err error) {
	if err != nil && err != sql.ErrNoRows {
		log.Error("CacheHandler: Failed to perform database operation")
		// FIXME client.dieWithError(e);
		panic(err)
	}
}

func NewDB(db string) (h *DB, err error) {
	h = &DB{}

	log.Info("CacheHandler: Loading database from " + db)

	sqlite, err := sql.Open("sqlite3", db)
	if err != nil {
		return
	}
	// defer db.Close()

	h.sqlite = sqlite

	//try {

	//Class.forName("org.sqlite.JDBC");
	//sqlite = DriverManager.getConnection("jdbc:sqlite:" + db);
	//DatabaseMetaData dma = sqlite.getMetaData();
	//Out.info("CacheHandler: Using " + dma.getDatabaseProductName() + " " + dma.getDatabaseProductVersion() + " over " + dma.getDriverName() + " " + dma.getJDBCMajorVersion() + "." + dma.getJDBCMinorVersion() + " running in " + dma.getDriverVersion() + " mode");

	log.Info("CacheHandler: Initializing database tables...")
	//Statement stmt = sqlite.createStatement();
	sqlite.Exec(`CREATE TABLE IF NOT EXISTS CacheList (
		fileid VARCHAR(65) NOT NULL,
		lasthit INT UNSIGNED NOT NULL,
		filesize INT UNSIGNED NOT NULL,
		active BOOLEAN NOT NULL,
		PRIMARY KEY(fileid)
	);`)
	sqlite.Exec("CREATE INDEX IF NOT EXISTS Lasthit ON CacheList (lasthit DESC);")
	sqlite.Exec(`CREATE TABLE IF NOT EXISTS StringVars (
		k VARCHAR(255) NOT NULL,
		v VARCHAR(255) NOT NULL,
		PRIMARY KEY(k)
	);`)

	h.cacheIndexClearActive, _ = sqlite.Prepare("UPDATE CacheList SET active=0;")
	h.cacheIndexCountStats, _ = sqlite.Prepare("SELECT COUNT(*), SUM(filesize) FROM CacheList;")
	h.queryCachelistSegment, _ = sqlite.Prepare("SELECT fileid FROM CacheList WHERE fileid BETWEEN ? AND ?;")
	h.queryCachedFileLasthit, _ = sqlite.Prepare("SELECT lasthit FROM CacheList WHERE fileid=?;")
	h.queryCachedFileSortOnLasthit, _ = sqlite.Prepare(`SELECT fileid, lasthit, filesize
		FROM CacheList ORDER BY lasthit LIMIT ?, ?;`)
	h.insertCachedFile, _ = sqlite.Prepare(`INSERT OR REPLACE INTO CacheList
		(fileid, lasthit, filesize, active) VALUES (?, ?, ?, 1);`)
	h.updateCachedFileActive, _ = sqlite.Prepare("UPDATE CacheList SET active=1 WHERE fileid=?;")
	h.updateCachedFileLasthit, _ = sqlite.Prepare("UPDATE CacheList SET lasthit=? WHERE fileid=?;")
	h.deleteCachedFile, _ = sqlite.Prepare("DELETE FROM CacheList WHERE fileid=?;")
	h.deleteCachedFileInactive, _ = sqlite.Prepare("DELETE FROM CacheList WHERE active=0;")
	h.setStringVar, _ = sqlite.Prepare("INSERT OR REPLACE INTO StringVars (k, v) VALUES (?, ?);")
	h.getStringVar, _ = sqlite.Prepare("SELECT v FROM StringVars WHERE k=?;")

	//try {
	// convert and clear pre-r81 tablespace if present.
	// this will trip an exception if the table doesn't exist and skip the rest of the conversion block
	sqlite.Exec("UPDATE CacheIndex SET active=0;")

	log.Info("Updating database schema to r81...")
	hashtable := make(map[string]int64)
	// TODO
	//java.util.Hashtable<String, Long> hashtable = new java.util.Hashtable<String, Long>();
	//ResultSet rs = stmt.executeQuery("SELECT fileid, lasthit FROM CacheIndex;");
	//while(rs.next()) {
	//hashtable.put(rs.getString(1), new Long(rs.getLong(2)));
	//}
	//rs.close();

	//sqlite.setAutoCommit(false);

	for fileid, value := range hashtable {
		hvf, _ := NewHVFileFromId(fileid)
		h.insertCachedFile.Exec(fileid, value, hvf.Size())
	}

	//sqlite.setAutoCommit(true)

	sqlite.Exec("DROP TABLE CacheIndex;")

	log.Info("Database updates complete")
	//}
	//catch(Exception e) {}
	return

	return
}

func (h *DB) optimize() {
	log.Info("CacheHandler: Optimizing database...")
	h.sqlite.Exec("VACUUM;")
}

func (h *DB) terminate() {
	if h.sqlite != nil {
		h.setStringVar.Exec(CLEAN_SHUTDOWN_KEY, CLEAN_SHUTDOWN_VALUE)
		h.sqlite.Close()
		h.sqlite = nil
	}
}

func (h *DB) CleanShutdown() (val string, err error) {
	defer func() { h.check(err) }()
	err = h.getStringVar.QueryRow(CLEAN_SHUTDOWN_KEY).Scan(&val)
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (h *DB) SetCleanShutdown(val interface{}) (err error) {
	defer func() { h.check(err) }()
	_, err = h.setStringVar.Exec(CLEAN_SHUTDOWN_KEY, val)
	return
}

func (h *DB) InsertCachedFile(lasthit int64, hvFile *HVFile) (err error) {
	defer func() { h.check(err) }()
	_, err = h.insertCachedFile.Exec(hvFile.Id(), lasthit, hvFile.Size())
	return
}

func (h *DB) CountStats() (count int, size int64, err error) {
	defer func() { h.check(err) }()
	err = h.cacheIndexCountStats.QueryRow().Scan(&count, &size)
	/*if err == sql.ErrNoRows {
		err = nil
	}*/
	return
}

func (h *DB) CachedFilesSegment(segment string) (fileList []string, err error) {
	defer func() { h.check(err) }()

	rows, err := h.queryCachelistSegment.Query(segment+"0", segment+"g")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var fileid string
		if e := rows.Scan(&fileid); e != nil {
			err = e
			return
		}
		if !Settings.isStaticRange(fileid) {
			fileList = append(fileList, fileid)
		}
	}

	err = rows.Err()
	return
}

func (h *DB) CachedFileLasthit(id string) (hittime int64, err error) {
	defer func() { h.check(err) }()
	err = h.cacheIndexCountStats.QueryRow(id).Scan(&hittime)
	return
}

func (h *DB) SetCachedFileLasthit(lasthit int64, id string) (err error) {
	defer func() { h.check(err) }()
	_, err = h.cacheIndexCountStats.Exec(lasthit, id)
	return
}

func (h *DB) CachedFileSortOnLasthit(start, end int) (ids []string, hits []int64, err error) {
	defer func() { h.check(err) }()

	rows, err := h.queryCachedFileSortOnLasthit.Query(start, end)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var fileid string
		var hit int64
		if e := rows.Scan(&fileid, &hit); e != nil {
			err = e
			return
		}
		ids = append(ids, fileid)
		hits = append(hits, hit)
	}

	err = rows.Err()
	return
}

func (h *DB) DeleteCachedFile(id string) (err error) {
	defer func() { h.check(err) }()
	_, err = h.deleteCachedFile.Exec(id)
	return
}
func (h *DB) UpdateCachedFileActive(id string) (affected int64, err error) {
	defer func() { h.check(err) }()

	r, err := h.updateCachedFileActive.Exec(id)
	if err != nil {
		return
	}

	affected, err = r.RowsAffected()
	return
}
