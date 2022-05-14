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
	first  bool
	w      io.Writer
}

func (s *prefixingWriter) Write(p []byte) (int, error) {
	if !s.first {
		s.first = !s.first
		if _, err := s.w.Write(s.prefix); err != nil {
			return 0, err
		}
	}
	n := 0
	for {
		i := bytes.IndexByte(p, '\n')
		if i == -1 || len(p)-1 == i {
			if len(p)-1 == i {
				s.first = !s.first
			}
			nn, err := s.w.Write(p)
			n += nn
			if err != nil {
				return n, err
			}
			break
		}

		nn, err := s.w.Write(p[:i+1])
		n += nn
		if err != nil {
			return n, err
		}
		if _, err := s.w.Write(s.prefix); err != nil {
			return 0, err
		}
		p = p[i+1:]
	}
	return n, nil
}
