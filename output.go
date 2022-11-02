// Copyright 2022 Twitter, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func (s *prefixingWriter) Close() error {
	// TODO track double-closing?

	if s.buf.Len() == 0 {
		return nil
	}
	_, err := s.w.Write(s.buf.Bytes())
	if err != nil {
		return err
	}
	if (s.buf.Bytes())[s.buf.Len()-1] == '\n' {
		return nil
	}
	// cmd ended without trailing \n, add so that prefixed printing is not malformed
	_, err = s.w.Write([]byte{'\n'})
	return err
}
