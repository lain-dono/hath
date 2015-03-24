package cache

import (
	log "../log"
	"fmt"
	//"io"
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
)

// NOTE: this class does not necessarily represent an actual file even though it is occasionally
// used as such (getLocalFileRef()) - it is an abstract representation of files in the HentaiVerse System

type HVFile struct {
	hash             string
	size, xres, yres int
	_type            string
}

func (hvf *HVFile) LocalFileRef() (file *os.File, err error) {
	file, err = os.OpenFile(hvf.LocalFilePath(), os.O_RDONLY, 0)
	return
}
func (hvf *HVFile) LocalFilePath() string {
	return path.Join(handler.CacheDir(), hvf.hash[:2], hvf.Id())
}

/*

	public boolean localFileMatches(File file) {
		// NOTE: we only check the sha-1 hash and filesize here,
		//to save resources and avoid dealing with the crummy image handlers
		try {
			return file.length() == size && hash.startsWith(MiscTools.getSHAString(file));
		} catch(java.io.IOException e) {
			Out.warning("Failed reading file " + file + " to determine hash.");
		}

		return false;
	}
*/

// accessors
func (hvf *HVFile) MimeType() string {
	switch hvf._type {
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

func (hvf *HVFile) String() string { return hvf.Id() }
func (hvf *HVFile) Id() string {
	return fmt.Sprintf("%s-%d-%d-%d-%s", hvf.hash, hvf.size, hvf.xres, hvf.yres, hvf._type)
}

func (hvf *HVFile) Hash() string { return hvf.hash }
func (hvf *HVFile) Size() int64  { return int64(hvf.size) }
func (hvf *HVFile) Type() string { return hvf._type }

// static stuff

var _validHVId = regexp.MustCompile(`^[a-f0-9]{40}-[0-9]{1,8}-[0-9]{1,5}-[0-9]{1,5}-((jpg)|(png)|(gif))$`)

func IsValidId(fileid string) bool {
	return _validHVId.MatchString(fileid)
}

func NewHVFile(hash string, size, xres, yres int, t string) *HVFile {
	return &HVFile{
		hash:  hash,
		size:  size,
		xres:  xres,
		yres:  yres,
		_type: t,
	}
}

func NewHVFileFromFile(file *os.File, verify bool) (hvf *HVFile, err error) {
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
	*/
}

func NewHVFileFromId(fileid string) (hvf *HVFile, err error) {
	if !IsValidId(fileid) {
		err = fmt.Errorf(`Invalid fileid "%s"`, fileid)
		log.Warn(err)
		return
	}

	defer func() {
		if e := recover(); e != nil {
			hvf = nil
			err = fmt.Errorf(`Failed to parse fileid "%s" : %s`, fileid, e)
			log.Warn(err)
		}
	}()

	fileidParts := strings.Split(fileid, "-")
	hash := fileidParts[0]
	size, err := strconv.Atoi(fileidParts[1])
	xres, err := strconv.Atoi(fileidParts[2])
	yres, err := strconv.Atoi(fileidParts[3])
	_type := fileidParts[4]

	hvf = NewHVFile(hash, size, xres, yres, _type)
	return
}

func HVFIds(arr []HVFile) (ids []string) {
	for _, f := range arr {
		ids = append(ids, f.Id())
	}
	return
}
