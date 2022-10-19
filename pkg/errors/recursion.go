// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
	"reflect"
)

type recursionError struct {
	wrapped error
	kind    string
	elem    interface{}
	hist    []interface{}
}

// ErrRecusion describes a resursion errors caused by a dedicated element with an element history.
func ErrRecusion(kind string, elem interface{}, hist interface{}) error {
	return &recursionError{nil, kind, elem, ToInterfaceSlice(hist)}
}

func ErrRecusionWrap(err error, kind string, elem interface{}, hist interface{}) error {
	return &recursionError{err, kind, elem, ToInterfaceSlice(hist)}
}

func (e *recursionError) Error() string {
	msg := fmt.Sprintf("%s recursion: use of %v", e.kind, e.elem)
	if len(e.hist) > 0 {
		s := ""
		sep := ""
		for _, h := range e.hist {
			s = fmt.Sprintf("%s%s%v", s, sep, h)
			sep = "->"
		}
		msg = fmt.Sprintf("%s for %s", msg, s)
	}
	if e.wrapped != nil {
		return msg + ": " + e.wrapped.Error()
	}
	return msg
}

func (e *recursionError) Unwrap() error {
	return e.wrapped
}

func (e *recursionError) Elem() interface{} {
	return e.elem
}

func (e *recursionError) Kind() string {
	return e.kind
}

func IsErrRecusion(err error) bool {
	return IsA(err, &recursionError{})
}

func IsErrRecursionKind(err error, kind string) bool {
	var uerr *recursionError
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}

func ToInterfaceSlice(list interface{}) []interface{} {
	if list == nil {
		return nil
	}
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		panic("no array or slice")
	}
	r := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		r[i] = v.Index(i).Interface()
	}
	return r
}
