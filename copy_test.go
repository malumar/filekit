package filekit

import (
	"fmt"
	"github.com/malumar/shellexec"
	"github.com/malumar/tracer"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {

	if pth, err := os.MkdirTemp("", "test*"); err != nil {
		t.Error(err)
	} else {
		src := filepath.Join(pth, "file1")

		assert.NoError(t, Touch(src, 0644))
		if env, err := shellexec.NewEnvironmentFromProcess(); err != nil {
			t.Error(err)
		} else {
			tc := tracer.NewSimple(tracer.All, func(bytes []byte) {
				fmt.Println(string(bytes))
			})
			assert.NoError(t, ShellCopyWithTracer(tc, src, filepath.Join(pth, "file2"), 0644, env))

			assert.NoError(t, ShellPrintEnv(tc, env))
		}

	}

}
