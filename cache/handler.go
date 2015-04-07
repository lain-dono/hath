package cache

type Handler struct {
	cacheDir, tmpDir string
	DiskLimitBytes   int64
	*DB
}

func NewHandler(cacheDir, tmpDir, db string) (h *Handler) {
	return &Handler{
		cacheDir: cacheDir,
		tmpDir:   tmpDir,
		DB:       NewDB(db),
	}
}
