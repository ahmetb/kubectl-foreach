package main

import (
	"bytes"
	"io"
	"sync"
)

type synchronizedWriter struct {
	io.Writer
	sync.Mutex
}

func (s *synchronizedWriter) Write(p []byte) (int, error) {
	s.Lock()
	defer s.Unlock()
	return s.Writer.Write(p)
}

type prefixingWriter struct {
	prefix []byte
	w      io.Writer // has per-Write mutex

	buf bytes.Buffer
}

func (s *prefixingWriter) Write(p []byte) (int, error) {
	n := len(p)
	for {
		if s.buf.Len() == 0 {
			s.buf.Write(s.prefix)
		}

		i := bytes.IndexByte(p, '\n')
		if i == -1 {
			_, err := s.buf.Write(p)
			if err != nil {
				return 0, err
			}
			break
		}

		// found \n
		if _, err := s.buf.Write(p[:i+1]); err != nil {
			return 0, err
		}

		_, err := s.w.Write(s.buf.Bytes())
		if err != nil {
			return 0, err
		}
		s.buf.Reset()

		if i == len(p)-1 {
			break
		}
		p = p[i+1:]
	}
	return n, nil
}
