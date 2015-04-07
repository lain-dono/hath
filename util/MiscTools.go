package util

import (
	"crypto/sha1"
	"fmt"
	"io"
	//"hash"
)

func SHA(from string) (sum string) {
	return fmt.Sprintf("%x", sha1.Sum([]byte(from)))
}

func SHAreader(from io.Reader) (sum string, err error) {
	hasher := sha1.New()
	hasher.Reset()
	_, err = io.Copy(hasher, from)
	if err != nil {
		return
	}
	sum = fmt.Sprintf("%x", hasher.Sum(nil))
	return
}

/*
	// these two functions are used to process servercmd type GETs
	public static Hashtable<String,String> parseAdditional(String additional) {
		Hashtable<String,String> addTable = new Hashtable<String,String>();

		if(additional != null) {
			if(!additional.isEmpty()) {
				String[] keyValuePairs = additional.trim().split(";");

				for(String kvPair : keyValuePairs) {
					if(kvPair.length() > 2) {
						String[] kvPairParts = kvPair.trim().split("=", 2);

						if(kvPairParts.length == 2) {
							addTable.put(kvPairParts[0].trim(), kvPairParts[1].trim());
						}
						else {
							Out.warning("Invalid kvPair: " + kvPair);
						}
					}
				}
			}
		}

		return addTable;
	}
*/

//func SHAString(from string) string {
//return SHAStringBytes([]byte(from))
//}

//func SHABytes([]byte) {

//}

/*
	public static String getSHAString(File from) throws java.io.IOException {
		return getSHAString(FileTools.getFileContents(from));
	}

	public static String getSHAString(byte[] bytes) {
		java.lang.StringBuffer sb = null;

		try {
			java.security.MessageDigest md = java.security.MessageDigest.getInstance("SHA");
			byte[] keybytes = md.digest(bytes);
			sb = new java.lang.StringBuffer(keybytes.length * 2);

			for(byte b : keybytes) {
				String s = Integer.toHexString((int) b & 0xff);
				sb.append((s.length() < 2 ? "0" : "") + s);
			}

			// for some reason this doesn't appear to be releasing properly, so we'll do it manually...
			md.reset();
			md = null;
			bytes = null;
			keybytes = null;
		} catch(java.security.NoSuchAlgorithmException e) {
			HentaiAtHomeClient.dieWithError(e);
		}

		return sb == null ? null : sb.toString().toLowerCase();
	}


*/
