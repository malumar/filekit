package filekit

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"syscall"
	"time"
)

func IsFileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
func IsFileExistsAndIsDir(filename string) bool {
	fi, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	if fi.IsDir() {
		return true
	} else {
		return false
	}
}

func IsFileExistsAndIsFile(filename string) bool {
	fi, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	if !fi.IsDir() {
		return true
	} else {
		return false
	}
}

func IsSymlink(filename string) (error, bool) {
	fi, err := os.Lstat(filename)
	if err != nil {
		return err, false
	} else {
		if fi.Mode()&os.ModeSymlink != 0 {
			return nil, true
		}
		return nil, false
	}
}
func MkDirIfNotExists(path string, chmod os.FileMode) (err error) {
	if !IsFileExists(path) {
		if err = os.Mkdir(path, chmod); err != nil {
			// or maybe someone hasn't created one before
			if !os.IsExist(err) {
				return err
			}
		}

	}
	return nil
}

func IsSymlinkOrNot(filename string) bool {
	err, ok := IsSymlink(filename)
	return err == nil && ok
}

func Touch(filename string, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, perm)
	fmt.Println("##", err)
	if err != nil {
		return err
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func SaveToFileWithLock(filename string, timeOut time.Duration, fileMode os.FileMode, bytes []byte) (err error) {
	var file *os.File
	var locked bool
	defer func() {
		if file != nil {
			if locked {
				errl := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
				if errl != nil {
					Logger("LOCK_UN (write) %s [ERROR %v]", filename, errl)
				} else {
					Logger("LOCK_UN (write) %s [OK]", filename)
				}

			}
			file.Close()
		}
	}()

	if file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, fileMode); err != nil {
		return
	}
	var how int

	if timeOut > 0 {
		how = syscall.LOCK_EX + syscall.LOCK_NB
		Logger("LOCK_EX with timeout %s", filename)
	} else {
		Logger("LOCK_EX with block %s", filename)
		how = syscall.LOCK_EX
	}

	try := 0
	for {
		if err = syscall.Flock(int(file.Fd()), how); err != nil {
			if timeOut > 0 {
				if try == 0 {
					Logger("LOCK_EX Sleep %s", filename)
					try++
					time.Sleep(timeOut)
					continue
				}
				return err

			}
		}

		locked = true
		break

	}

	locked = true
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}
	err = file.Truncate(0)
	if err != nil {
		return
	}

	_, err = file.Write(bytes)
	return
}

var Logger = func(format string, args ...interface{}) {
	slog.Debug(fmt.Sprintf(format, args))
}
