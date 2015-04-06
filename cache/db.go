// +build ignore
package cache

import (
	log "../log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

const (
	CLEAN_SHUTDOWN_KEY   = "clean_shutdown"
	CLEAN_SHUTDOWN_VALUE = "clean_r81"
)

type CachedFile struct {
	Id      Id
	Lasthit time.Time
	Size    int64
	Active  bool
}

type DB struct {
	//isForceDirty bool

	*sql.DB

	cacheIndexClearActive, cacheIndexCountStats                       *sql.Stmt
	queryCachelistSegment                                             *sql.Stmt
	queryCachedFileLasthit, queryCachedFileSortOnLasthit              *sql.Stmt
	insertCachedFile, updateCachedFileLasthit, updateCachedFileActive *sql.Stmt
	deleteCachedFile, deleteCachedFileInactive                        *sql.Stmt
	getStringVar, setStringVar                                        *sql.Stmt

	staticRanges map[string]bool
}

func (h *DB) SetStaticRanges(ranges map[string]bool) {
	h.staticRanges = ranges
}
func (h *DB) IsStaticRanges(fileid string) (ok bool) {
	_, ok = h.staticRanges[fileid[0:4]]
	return
}

func (h *DB) check(err error) {
	if err != nil && err != sql.ErrNoRows {
		log.Error("CacheHandler: Failed to perform database operation")
		// FIXME client.dieWithError(e);
		panic(err)
	}
}

func (h *DB) mustPrepare(q string) (stmt *sql.Stmt) {
	stmt, err := h.Prepare(q)
	if err != nil {
		panic(err)
	}
	return
}
func (h *DB) mustExec(q string) {
	_, err := h.Exec(q)
	if err != nil {
		panic(err)
	}
}

func NewDB(db string) (h *DB) {
	h = &DB{
		staticRanges: make(map[string]bool),
	}
	//log.Info("CacheHandler: Loading database from " + db)

	sqlite, err := sql.Open("sqlite3", db)
	if err != nil {
		panic(err)
	}
	h.DB = sqlite

	//log.Info("CacheHandler: Initializing database tables...")

	h.initTables()
	h.initStmt()

	//log.Info("Updating database schema to r81...")
	err = h.initPre81() // TODO maybe ignore it
	if err != nil {
		panic(err)
	}
	//log.Info("Database updates complete")

	return
}

func (h *DB) initTables() {
	h.mustExec(`CREATE TABLE IF NOT EXISTS CacheList (
		fileid VARCHAR(65) NOT NULL,
		lasthit INT UNSIGNED NOT NULL,
		filesize INT UNSIGNED NOT NULL,
		active BOOLEAN NOT NULL,
		PRIMARY KEY(fileid)
	);`)
	h.mustExec("CREATE INDEX IF NOT EXISTS Lasthit ON CacheList (lasthit DESC);")
	h.mustExec(`CREATE TABLE IF NOT EXISTS StringVars (
		k VARCHAR(255) NOT NULL,
		v VARCHAR(255) NOT NULL,
		PRIMARY KEY(k)
	);`)
}

func (h *DB) initStmt() {
	h.cacheIndexClearActive = h.mustPrepare("UPDATE CacheList SET active=0;")
	h.cacheIndexCountStats = h.mustPrepare("SELECT COUNT(*), SUM(filesize) FROM CacheList;")
	h.queryCachelistSegment = h.mustPrepare("SELECT fileid FROM CacheList WHERE fileid BETWEEN ? AND ?;")
	h.queryCachedFileLasthit = h.mustPrepare("SELECT lasthit FROM CacheList WHERE fileid=?;")
	h.queryCachedFileSortOnLasthit = h.mustPrepare(`SELECT fileid, lasthit, filesize, active
		FROM CacheList ORDER BY lasthit LIMIT ?, ?;`) // FIXME remove active
	h.insertCachedFile = h.mustPrepare(`INSERT OR REPLACE INTO CacheList
		(fileid, lasthit, filesize, active) VALUES (?, ?, ?, 1);`)
	h.updateCachedFileActive = h.mustPrepare("UPDATE CacheList SET active=1 WHERE fileid=?;")
	h.updateCachedFileLasthit = h.mustPrepare("UPDATE CacheList SET lasthit=? WHERE fileid=?;")
	h.deleteCachedFile = h.mustPrepare("DELETE FROM CacheList WHERE fileid=?;")
	h.deleteCachedFileInactive = h.mustPrepare("DELETE FROM CacheList WHERE active=0;")
	h.setStringVar = h.mustPrepare("INSERT OR REPLACE INTO StringVars (k, v) VALUES (?, ?);")
	h.getStringVar = h.mustPrepare("SELECT v FROM StringVars WHERE k=?;")
}

// convert and clear pre-r81 tablespace if present.
// this will trip an exception if the table doesn't exist and skip the rest of the conversion block
func (h *DB) initPre81() (err error) {
	var exists int
	err = h.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='CacheIndex';").Scan(&exists)
	if err != nil || exists == 0 {
		return
	}

	_, err = h.Exec("UPDATE CacheIndex SET active=0;")
	if err != nil {
		return
	}

	rows, err := h.Query("SELECT fileid, lasthit FROM CacheIndex;")
	if err != nil {
		return
	}
	defer rows.Close()

	var fileid string
	var lasthit int64
	for rows.Next() {
		err = rows.Scan(&fileid, &lasthit)
		if err != nil {
			return
		}

		id := mustId(fileid)
		err = h.InsertCachedFile(id, time.Unix(lasthit, 0))
		if err != nil {
			return
		}
	}

	err = rows.Err()
	if err != nil {
		return
	}

	_, err = h.Exec("DROP TABLE CacheIndex;")
	return
}

func (h *DB) Optimize() {
	//log.Info("CacheHandler: Optimizing database...")
	h.mustExec("VACUUM;")
}

func (h *DB) Terminate() {
	if h.DB != nil {
		h.setStringVar.Exec(CLEAN_SHUTDOWN_KEY, CLEAN_SHUTDOWN_VALUE)
		h.DB.Close()
		h.DB = nil
	}
}

// ------------------------------

func (h *DB) ClearActive() (n int64, err error) {
	res, err := h.cacheIndexClearActive.Exec()
	if err != nil {
		return
	}
	n, err = res.RowsAffected()
	return
}
func (h *DB) Activate(id Id) (err error) {
	_, err = h.updateCachedFileActive.Exec(id.String())
	return
}
func (h *DB) SetLastHit(id Id, t time.Time) (err error) {
	nix := t.Unix()
	_, err = h.updateCachedFileLasthit.Exec(nix, id.String())
	return
}

func (h *DB) CountStats() (count int, size int64, err error) {
	err = h.cacheIndexCountStats.QueryRow().Scan(&count, &size)
	return
}
func (h *DB) LastHit(id Id) (t time.Time, err error) {
	var val int64
	err = h.queryCachedFileLasthit.QueryRow(id.String()).Scan(&val)
	t = time.Unix(val, 0)
	return
}
func (h *DB) Segment(from, to string) (ids []Id, err error) {
	rows, err := h.queryCachelistSegment.Query(from, to)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var fileid string
		err = rows.Scan(&fileid)
		if err != nil {
			return
		}

		if !h.IsStaticRanges(fileid) {
			ids = append(ids, mustId(fileid))
		}
	}
	err = rows.Err()
	return
}

func (h *DB) CachedFileSortOnLasthit(from, count int) (files []CachedFile, err error) {
	rows, err := h.queryCachedFileSortOnLasthit.Query(from, count)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var fileid string
		var lasthit, filesize int64
		var active bool

		/*
			err = h.queryCachedFileSortOnLasthit.QueryRow(id.String()).Scan(&fileid, &lasthit, &filesize, &active)
			if err != nil {
				return
			}
		*/

		err = rows.Scan(&fileid, &lasthit, &filesize, &active)
		if err != nil {
			return
		}
		file := CachedFile{
			Id:      mustId(fileid),
			Lasthit: time.Unix(lasthit, 0),
			Size:    filesize,
			Active:  active,
		}

		files = append(files, file)
	}
	err = rows.Err()

	return
}

func (h *DB) InsertCachedFile(id Id, lasthit time.Time) (err error) {
	_, err = h.insertCachedFile.Exec(id.String(), lasthit.Unix(), id.Size())
	return
}

func (h *DB) Remove(id Id) (err error) {
	_, err = h.deleteCachedFile.Exec(id.String())
	return
}
func (h *DB) RemoveInactive() (n int64, err error) {
	res, err := h.deleteCachedFileInactive.Exec()
	if err != nil {
		return
	}
	n, err = res.RowsAffected()
	return
}

// ------------------------------

func (h *DB) CleanShutdown() (val string, err error) {
	//defer func() { h.check(err) }()
	err = h.getStringVar.QueryRow(CLEAN_SHUTDOWN_KEY).Scan(&val)
	//if err == sql.ErrNoRows {
	//err = nil
	//}
	return
}

func (h *DB) SetCleanShutdown(val interface{}) (err error) {
	//defer func() { h.check(err) }()
	_, err = h.setStringVar.Exec(CLEAN_SHUTDOWN_KEY, val)
	return
}

/*
func (h *DB) InsertCachedFile(lasthit int64, hvFile *HVFile) (err error) {
	defer func() { h.check(err) }()
	_, err = h.insertCachedFile.Exec(hvFile.Id(), lasthit, hvFile.Size())
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
*/
