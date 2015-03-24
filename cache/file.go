package cache

type CachedFile struct {
	id        string
	needFlush bool
}

func NewCachedFile(fileid string) *CachedFile {
	return &CachedFile{
		id:        fileid,
		needFlush: false,
	}
}

func (file *CachedFile) Id() string               { return file.id }
func (file *CachedFile) HVFile() (*HVFile, error) { return NewHVFileFromId(file.id) }
func (file *CachedFile) NeedsFlush() bool         { return file.needFlush }
func (file *CachedFile) Flushed()                 { file.needFlush = false }
func (file *CachedFile) Hit() {
	file.needFlush = true

	//synchronized(recentlyAccessed) {
	//needFlush = true;
	//recentlyAccessed.add(this);
	//}
}
