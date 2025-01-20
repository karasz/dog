// Package multireadcloser provides utilities for working with multiple io.ReadCloser instances.
// It allows combining multiple io.ReadCloser objects into a single io.ReadCloser,
// enabling sequential reading and closing of multiple readers.
// Most of it is a copy of the io.MultiReader implementation from the Go standard library.
package multireadcloser

import "io"

type eofReadCloser struct{}

func (eofReadCloser) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (eofReadCloser) Close() error {
	return nil
}

type multiReadCloser struct {
	readers []io.ReadCloser
}

// MultiReadCloser combines multiple io.ReadCloser instances into a single
// io.ReadCloser. It returns a new io.ReadCloser that reads sequentially
// from each provided reader until all are exhausted. The returned
// io.ReadCloser also closes all underlying readers when closed.
func MultiReadCloser(readers ...io.ReadCloser) io.ReadCloser {
	r := make([]io.ReadCloser, len(readers))
	copy(r, readers)
	return &multiReadCloser{r}
}

//revive:disable:cognitive-complexity
func (mr *multiReadCloser) Read(p []byte) (n int, err error) {
	//revive:enable:cognitive-complexity
	for len(mr.readers) > 0 {
		// Optimization to flatten nested multiReaders (Issue 13558).
		if len(mr.readers) == 1 {
			if r, ok := mr.readers[0].(*multiReadCloser); ok {
				mr.readers = r.readers
				continue
			}
		}
		n, err = mr.readers[0].Read(p)
		if err == io.EOF {
			// Use eofReader instead of nil to avoid nil panic
			// after performing flatten (Issue 18232).
			mr.readers[0] = eofReadCloser{} // permit earlier GC
			mr.readers = mr.readers[1:]
		}
		if n > 0 || err != io.EOF {
			if err == io.EOF && len(mr.readers) > 0 {
				// Don't return EOF yet. More readers remain.
				err = nil
			}
			return n, err
		}
	}
	return 0, io.EOF
}
func (mr *multiReadCloser) Close() error {
	errlist := make([]error, 0, len(mr.readers))
	for _, r := range mr.readers {
		e := r.Close()
		if e != nil {
			errlist = append(errlist, e)
		}
	}
	if len(errlist) == 0 {
		return nil
	}
	return errlist[0] // Return the first error for now, TODO: return all errors
}
