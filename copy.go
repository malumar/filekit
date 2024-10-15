package filekit

import (
	"fmt"
	"github.com/malumar/shellexec"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

const (
	cmdCopy  = "/usr/bin/cp"
	cmdShell = "/usr/bin/sh"
)

type CopyMode uint32

const (
	CM_FORCE                  CopyMode = 1 << (32 - 1 - iota) // 1 << 0 which is 0000000000000001
	CM_RECURSIVE                                              // 1 << 1 which is 0000000000000010
	CM_ARCHIVE                                                // 1 << 2 which is 0000000000000100
	CM_FOLLOW_SYMLINK                                         // 1 << 3 which is 0000000000001000
	CM_HARD_LINK                                              // 1 << 4 witch is 0000000000010000
	CM_SYMBOLICK_LINK                                         // 1 << 5 witch is 0000000000100000
	CM_UPDATE                                                 // 1 << 6 witch is 0000000001000000
	CM_VERBOSE                                                // 1 << 7 witch is 0000000010000000
	CM_STAY_FILE_SYSTEM                                       // 1 << 8 witch is 0000000100000000
	CM_STRIP_TRAILING_SLASHES                                 // 1 << 9 witch is 0000001000000000
	CM_PARENTS                                                // 1 <<10 witch is 0000010000000000
	CM_NO_DEREFERENCE                                         // 1 <<11 witch is 0000100000000000
	CM_DO_NOT_OVERWRITE                                       // 1 <<12 witch is 0001000000000000
	CM_DEREFERENCE                                            // 1 <<13 witch is 0010000000000000
)

type Environment = map[string]string

func NewEnvironmentFromProcess() (Environment, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	return NewEnvironment(currentUser.Name, currentUser.HomeDir), nil
}

func NewEnvironment(login, home string) Environment {
	return Environment{"LOGNAME": login, "HOME": home}
}

func Copy(source string, destination string, mode CopyMode, env Environment) (err error) {
	return ShellCopyWithTracer(nil, source, destination, mode, env)
}

func ShellPrintEnv(tracer io.Writer, env Environment) (err error) {
	return shellexec.New(tracer).SetEnv(env).Run("/usr/bin/printenv").GoAndCleanup()
}

// ShellCopyWithTracer
// @env must have set LOGNAME & HOME
func ShellCopyWithTracer(tracer io.Writer, source string, destination string, mode CopyMode, env Environment) (err error) {
	args := make([]interface{}, 0, 17)
	if mode&CM_FORCE != 0 {
		args = append(args, "--force")

	}
	if mode&CM_ARCHIVE != 0 {
		args = append(args, "--archive")
	}
	if mode&CM_RECURSIVE != 0 {
		args = append(args, "-r")
	}

	if mode&CM_DO_NOT_OVERWRITE != 0 {
		args = append(args, "--no-clobber")
	}

	if mode&CM_HARD_LINK != 0 {
		args = append(args, "-l")
	}
	if mode&CM_DEREFERENCE != 0 {
		args = append(args, "-L")
	}
	if mode&CM_SYMBOLICK_LINK != 0 {
		args = append(args, "--symbolic-link")
	}
	if mode&CM_UPDATE != 0 {
		args = append(args, "--update")
	}
	if mode&CM_VERBOSE != 0 {
		args = append(args, "--verbose")
	}
	if mode&CM_FOLLOW_SYMLINK != 0 {
		args = append(args, "-H")
	}
	if mode&CM_STAY_FILE_SYSTEM != 0 {
		args = append(args, "--one-file-system")
	}
	if mode&CM_STRIP_TRAILING_SLASHES != 0 {
		args = append(args, "--strip-trailing-slashes")
	}
	if mode&CM_PARENTS != 0 {
		args = append(args, "--parents")
	}
	if mode&CM_NO_DEREFERENCE != 0 {
		args = append(args, "--no-dereference")
	}

	args = append(args, source, destination)

	return shellexec.New(tracer).SetEnv(env).Run(cmdCopy, args...).GoAndCleanup()

}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string, mode CopyMode) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	if IsFileExists(dst) && !IsSetUint32Flag(mode, CM_FORCE) {

		return fmt.Errorf("file %s already exists", dst)
	}

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}
func IsSetUint32Flag(mode CopyMode, flag CopyMode) bool {
	return mode&flag != 0
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string, mode CopyMode) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {

		if !IsSetUint32Flag(mode, CM_FORCE) {
			return fmt.Errorf("destination already exists")
		}

		if err = os.Chmod(dst, si.Mode()); err != nil {
			return
		}

	} else {
		err = os.MkdirAll(dst, si.Mode())
		if err != nil {
			return
		}

	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if IsSetUint32Flag(mode, CM_RECURSIVE) {

				err = CopyDir(srcPath, dstPath, mode)
				if err != nil {
					return
				}

			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath, mode)
			if err != nil {
				return
			}
		}
	}

	return
}
