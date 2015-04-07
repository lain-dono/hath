package cache

import (
	"../util"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var (
	CONTENT_TYPE_DEFAULT = "text/html; charset=iso-8859-1"
	CONTENT_TYPE_OCTET   = "application/octet-stream"
	CONTENT_TYPE_JPG     = "image/jpeg"
	CONTENT_TYPE_PNG     = "image/png"
	CONTENT_TYPE_GIF     = "image/gif"

	IdFormat = regexp.MustCompile(`^[a-f0-9]{40}-[0-9]{1,8}-[0-9]{1,5}-[0-9]{1,5}-((jpg)|(png)|(gif))$`)
)

// NOTE: this class does not necessarily represent an actual file even though it is occasionally
// used as such (getLocalFileRef()) - it is an abstract representation of files in the HentaiVerse System

type Id struct {
	hash             string
	size, xres, yres int
	ext              string
}

func (id *Id) Hash() string    { return id.hash }
func (id *Id) Size() int64     { return int64(id.size) }
func (id *Id) Ext() string     { return id.ext }
func (id *Id) Res() (int, int) { return id.xres, id.yres }

func (id *Id) String() string {
	return fmt.Sprintf("%s-%d-%d-%d-%s", id.hash, id.size, id.xres, id.yres, id.ext)
}

func (id *Id) MimeType() string {
	switch id.ext {
	case "jpg":
		return CONTENT_TYPE_JPG
	case "png":
		return CONTENT_TYPE_PNG
	case "gif":
		return CONTENT_TYPE_GIF
	default:
		return CONTENT_TYPE_OCTET
	}
}

func NewId(hash string, size, xres, yres int, ext string) Id {
	return Id{
		hash: hash,
		size: size,
		xres: xres,
		yres: yres,
		ext:  ext,
	}
}

func IsValidId(id string) bool { return IdFormat.MatchString(id) }

func NewIdFromString(s string) (id Id, ok bool) {
	ok = IsValidId(s)
	if !ok {
		return
	}
	id = mustId(s)
	return
}
func mustId(s string) Id {
	parts := strings.SplitN(s, "-", 5)
	hash := parts[0]
	size, _ := strconv.Atoi(parts[1])
	xres, _ := strconv.Atoi(parts[2])
	yres, _ := strconv.Atoi(parts[3])
	ext := parts[4]
	return NewId(hash, size, xres, yres, ext)
}

// Handler

func (h *Handler) IsExist(id Id) bool {
	_, err := os.Stat(h.ResolveId(id))
	return os.IsExist(err)
}

func (h *Handler) OpenFileById(id Id) (*os.File, error) {
	return os.OpenFile(h.ResolveId(id), os.O_RDONLY, 0)
}

func (h *Handler) ResolveId(id Id) string {
	return path.Join(h.cacheDir, id.hash[:2], id.String())
}

// NOTE: we only check the sha-1 hash and filesize here,
// to save resources and avoid dealing with the crummy image handlers
func (h *Handler) CheckFile(id Id) (ok bool, err error) {
	f, err := h.OpenFileById(id)
	if err != nil {
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil || stat.Size() != id.Size() {
		return
	}

	sum, err := util.SHAreader(f)
	ok = sum == id.Hash()
	return
}

/*

	public boolean localFileMatches(File file) {
		try {
			return file.length() == size && hash.startsWith(MiscTools.getSHAString(file));
		} catch(java.io.IOException e) {
			Out.warning("Failed reading file " + file + " to determine hash.");
		}

		return false;
	}
*/
/*
func (h *Handler) HVFileById(id string) (hvf HVFile, err error) {
	if !IsValidId(id) {
		err = fmt.Errorf(`Invalid fileid "%s"`, fileid)
		log.Warn(err)
		return
	}

	defer func() {
		if e := recover(); e != nil {
			hvf = nil
			err = fmt.Errorf(`Failed to parse fileid "%s" : %s`, id, e)
			log.Warn(err)
		}
	}()

	parts := strings.SplitN(id, "-")
	hash := parts[0]
	size, _ := strconv.Atoi(parts[1])
	xres, _ := strconv.Atoi(parts[2])
	yres, _ := strconv.Atoi(parts[3])
	_type := parts[4]

	hvf = NewHVFile(hash, size, xres, yres, _type)
	return
}

func NewHVFileFromFile(file *os.File, verify bool) (hvf HVFile, err error) {
	fileid := file.Name()
	hvf, err = NewHVFileFromId(fileid)
	return
	/*
		if(file.exists()) {
			String fileid = file.getName();

			try {
				if(verify) {
					if(!fileid.startsWith(MiscTools.getSHAString(file))) {
						return null;
					}
				}

				return getHVFileFromId(fileid);
			} catch(java.io.IOException e) {
				e.printStackTrace();
				Out.warning("Warning: Encountered IO error computing the hash value of " + file);
			}
		}

		return null;
	* /
}*/

/*
func NewHVFileFromId(fileid string) (hvf HVFile, err error) {
}
*/

// util
func Ids(arr []Id) (ids []string) {
	for _, f := range arr {
		ids = append(ids, f.String())
	}
	return
}
