// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package csv

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

// A Marshaler specifies directly how the data is marshaled to CSV.
// The result should include the header as well as the data.
type Marshaler interface {
	MarshalCSV() ([]byte, error)
}

// A Recorder specifies the Header that a file returns, as well as the string
// representation of the values. The lengths of elements returned by Header()
// and Record() should be the same.
//
// The reason for implementing Recorder is that a slice or array of Recorder
// can be marshaled.
type Recorder interface {
	Header() []string
	Record() []string
}

var (
	marshalerType = reflect.TypeOf(new(Marshaler)).Elem()
	recorderType  = reflect.TypeOf(new(Recorder)).Elem()
)

// Marshal takes a type that implements Marshaler, Recorder, or a slice or array
// of types that implement Recorder and returns a CSV byte slice.
//
// Even a slice of interface{} can be marshaled, provided that the type of every
// value in the slice is the same, and each type implements Recorder.
func Marshal(v interface{}) ([]byte, error) {
	vt := reflect.TypeOf(v)

	// Check right away if it implements Marshaler or Recorder.
	if vt.Implements(marshalerType) {
		t := v.(Marshaler)
		return t.MarshalCSV()
	}
	if vt.Implements(recorderType) {
		t := v.(Recorder)
		return marshalRecorder(t), nil
	}

	// Any of the other checks only make sense on non-pointers.
	if vt.Kind() == reflect.Ptr {
		return Marshal(reflect.ValueOf(v).Elem().Interface())
	}

	// We also support a slice or array of Recorder, but not of MarshalCSV, because
	// semantically, anything could be in MarshalCSV, especially the header, and
	// we don't want to try to guess.
	if vt.Kind() == reflect.Slice || vt.Kind() == reflect.Array {
		// Even if the slice has element type *Type, the following should still work.
		if vt.Elem().Implements(recorderType) {
			return marshalRecorderSlice(v)
		}

		// We might get a slice of some kind of interface, in which case we require that
		// each element is of the same type.
		if vt.Elem().Kind() == reflect.Interface {
			return marshalInterfaceSlice(v)
		}

		return nil, fmt.Errorf("csv: slice element type %s does not implement Recorder", vt.Elem())
	}

	return nil, fmt.Errorf("csv: cannot marshal %s", vt)
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
	if n == 0 {
		return nil, errors.New("csv: no data")
	}

	get := func(i int) Recorder {
		return vv.Index(i).Interface().(Recorder)
	}

	var buf bytes.Buffer
	writeRecord(&buf, get(0).Header())
	for i := 0; i < n; i++ {
		writeRecord(&buf, get(i).Record())
	}
	return buf.Bytes(), nil
}

// marshalInterfaceSlice takes a slice or array of an interface type.
//
// We require that each type in the slice/array is of the same type.
func marshalInterfaceSlice(v interface{}) (bs []byte, err error) {
	vv := reflect.ValueOf(v)
	n := vv.Len()
	if n == 0 {
		return nil, errors.New("csv: no data")
	}

	t := reflect.TypeOf(vv.Index(0).Interface())
	get := func(i int) Recorder {
		r, ok := vv.Index(i).Interface().(Recorder)
		if !ok {
			panic(fmt.Errorf("csv: slice element %T does not implement Recorder", vv.Index(i).Interface()))
		}
		if rt := reflect.TypeOf(r); rt != t {
			panic(fmt.Errorf("csv: expecting slice element type %s, got %s", t, rt))
		}
		return r
	}

	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
				return
			}
			// It's not an error, so continue panicking.
			panic(r)
		}
	}()

	var buf bytes.Buffer
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
