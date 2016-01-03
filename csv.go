// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package csv

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

type Marshaler interface {
	MarshalCSV() ([]byte, error)
}

type Recorder interface {
	Header() []string
	Record() []string
}

var (
	marshalerType = reflect.TypeOf(new(Marshaler)).Elem()
	recorderType  = reflect.TypeOf(new(Recorder)).Elem()
)

func Marshal(v interface{}) ([]byte, error) {
	vt := reflect.TypeOf(v)

	switch vt.Kind() {
	case reflect.Ptr:
		return Marshal(reflect.ValueOf(v).Elem().Interface())
	case reflect.Struct:
		if vt.Implements(marshalerType) {
			t := v.(Marshaler)
			return t.MarshalCSV()
		}
		if vt.Implements(recorderType) {
			t := v.(Recorder)
			return marshalRecorder(t), nil
		}
		return nil, fmt.Errorf("csv: struct type %s does not implement Marshaler or Recorder", vt)
	case reflect.Slice, reflect.Array:
		if vt.Elem().Kind() == reflect.Ptr {
			vt = vt.Elem() // now vt is a pointer
		}
		if vt.Elem().Implements(recorderType) {
			return marshalRecorderSlice(v)
		}
		return nil, fmt.Errorf("csv: slice element type %s does not implement Recorder", vt.String())
	default:
		return nil, fmt.Errorf("csv: cannot marshal type %s", vt)
	}
}

func marshalRecorder(v Recorder) []byte {
	var buf bytes.Buffer
	writeRecord(&buf, v.Header())
	writeRecord(&buf, v.Record())
	return buf.Bytes()
}

func marshalRecorderSlice(v interface{}) ([]byte, error) {
	vv := reflect.ValueOf(v)
	n := vv.Len()
	get := func(i int) Recorder {
		return vv.Index(i).Interface().(Recorder)
	}

	var buf bytes.Buffer
	if n == 0 {
		return nil, errors.New("csv: no data")
	}
	writeRecord(&buf, get(0).Header())
	for i := 0; i < n; i++ {
		writeRecord(&buf, get(i).Record())
	}
	return buf.Bytes(), nil
}

func writeRecord(buf *bytes.Buffer, slice []string) {
	m := len(slice) - 1
	for _, s := range slice[:m] {
		buf.WriteString(s)
		buf.WriteRune(',')
	}
	buf.WriteString(slice[m])
	buf.WriteRune('\n')
}
