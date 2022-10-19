// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	New    = errors.New
	Unwrap = errors.Unwrap
	Is     = errors.Is
	As     = errors.As
)

func Newf(msg string, args ...interface{}) error {
	return New(fmt.Sprintf(msg, args...))
}

func IsA(err error, target error) bool {
	if err == nil {
		return false
	}
	typ := reflect.TypeOf(target)

	for err != nil {
		if reflect.TypeOf(err).AssignableTo(typ) {
			return true
		}
		err = Unwrap(err)
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////

type wrappedError struct {
	wrapped error
	msg     string
}

func Wrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &wrappedError{
		wrapped: err,
		msg:     msg,
	}
}

func (e *wrappedError) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.wrapped)
}

func (e *wrappedError) Unwrap() error {
	return e.wrapped
}

// var errorType = reflect.TypeOf((*error)(nil)).Elem()

////////////////////////////////////////////////////////////////////////////////

type errinfo struct {
	wrapped error
	format  ErrorFormatter
	kind    string
	elem    *string
	ctx     string
}

func wrapErrInfo(err error, fmt ErrorFormatter, spec ...string) errinfo {
	e := newErrInfo(fmt, spec...)
	e.wrapped = err
	return e
}

func newErrInfo(fmt ErrorFormatter, spec ...string) errinfo {
	e := errinfo{
		format: fmt,
	}

	if len(spec) > 2 {
		e.kind = spec[0]
		e.elem = &spec[1]
		e.ctx = spec[2]
		return e
	}
	if len(spec) > 1 {
		e.kind = spec[0]
		e.elem = &spec[1]
		return e
	}
	if len(spec) > 0 {
		e.elem = &spec[0]
	}
	return e
}

func (e *errinfo) Error() string {
	msg := e.format.Format(e.kind, e.elem, e.ctx)
	if e.wrapped != nil {
		return msg + ": " + e.wrapped.Error()
	}
	return msg
}

func (e *errinfo) Unwrap() error {
	return e.wrapped
}

func (e *errinfo) Elem() *string {
	return e.elem
}

func (e *errinfo) Kind() string {
	return e.kind
}

func (e *errinfo) Ctx() string {
	return e.ctx
}

type Kinded interface {
	Kind() string
	SetKind(string)
}

func (e *errinfo) SetKind(kind string) {
	e.kind = kind
}
