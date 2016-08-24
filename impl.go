// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package csv

type marshaler struct {
	bs  []byte
	err error
}

func NewMarshaler(bs []byte, err error) Marshaler { return &marshaler{bs, err} }
func (m *marshaler) MarshalCSV() ([]byte, error)  { return m.bs, m.err }

type recorder struct {
	header []string
	record []string
}

func NewRecorder(h, r []string) Recorder { return &recorder{h, r} }
func (r *recorder) Header() []string     { return r.header }
func (r *recorder) Record() []string     { return r.record }
