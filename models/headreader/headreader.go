package headreader

import (
	"io"
	"os"
)

// New returns a headReader which reads first n bytes of a file
func New(f *os.File, n int64) io.ReadCloser {
	r := io.LimitReader(f, n)
	return &headReader{r, f}
}

type headReader struct {
	r io.Reader
	f *os.File
}

// Read implements the io.Reader interface
func (hr *headReader) Read(b []byte) (int, error) {
	return hr.r.Read(b)
}

// Close implements the io.Closer interface
func (hr *headReader) Close() error {
	return hr.f.Close()
}
