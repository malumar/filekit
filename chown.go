package filekit

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

func Chown(filename string, ownerUser, ownerGroup string) error {
	uid := -1
	gid := -1

	if strings.TrimSpace(ownerUser) != "" {
		if uidVal, err := strconv.Atoi(ownerUser); err == nil {
			// wygląda na to iż mamy liczbę
			uid = uidVal
		} else {
			if lookup, err := user.Lookup(ownerUser); err == nil {
				if uidVal, err := strconv.Atoi(lookup.Uid); err == nil {
					uid = uidVal
				} else {
					return fmt.Errorf("An error occurred while trying to convert UID=%v to int info about user `%v`: %v", lookup.Uid, ownerUser, err.Error())
				}
			} else {
				return fmt.Errorf("An error occurred while trying to retrieve information about user `%v`: %v", ownerUser, err.Error())

			}
		}
	}

	if strings.TrimSpace(ownerGroup) != "" {
		if gidVal, err := strconv.Atoi(ownerGroup); err == nil {
			// wygląda na to iż mamy liczbę
			gid = gidVal
		} else {
			if lookup, err := user.LookupGroup(ownerGroup); err == nil {
				if gidVal, err := strconv.Atoi(lookup.Gid); err == nil {
					gid = gidVal
				} else {
					return fmt.Errorf("An error occurred while trying to convert GID=%v to int info about group `%s`", lookup.Gid, err.Error())
				}
			} else {
				return fmt.Errorf("An error occurred while trying to retrieve info about group `%s`", err.Error())
			}
		}

	}

	if (uid == -1 && gid > -1) || (gid == -1 && uid > -1) {
		return fmt.Errorf("You must provide owner and group for changing the owner, you cannot provide only one parameter")
	} else {
		if err := os.Chown(filename, uid, gid); err != nil {
			return fmt.Errorf("Error `%v` occurred while trying to change the owner of file `%s:%s` (%d:%d)",
				err, ownerUser, ownerGroup, uid, gid)
		}
	}

	return nil

}

func Chmod(filename, perm string) error {
	if perm == "" {
		return nil
	}
	//fp := os.FileMode(converter.Oct2uint32(perm))
	//if err := os.Chmod(filename, fp); err != nil {
	//	return fmt.Errorf("chmod on %s (%v) failed: %v", perm, fp, err.Error())
	//}

	return nil
}
