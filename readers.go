package filekit

import (
	"bufio"
	"github.com/malumar/tracer"
	"os"
)

func ReadLineByLine(t tracer.Tracer, filename string, cb func(output []byte) bool) error {
	file, err := os.Open(filename)
	if err != nil {
		t.Error("file open error", "filename", filename, "err", err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if !cb(scanner.Bytes()) {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		t.Error("scanner error", "filename", filename, "error", err)
	}
	return err
}
