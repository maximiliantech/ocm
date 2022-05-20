// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package common

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/open-component-model/ocm/pkg/errors"
)

////////////////////////////////////////////////////////////////////////////////

type HistorySource interface {
	GetHistory() History
}

type History []NameVersion

func (h History) String() string {
	s := ""
	sep := ""
	for _, e := range h {
		s = fmt.Sprintf("%s%s%s", s, sep, e)
		sep = "->"
	}
	return s
}

func (h History) Contains(nv NameVersion) bool {
	for _, e := range h {
		if e == nv {
			return true
		}
	}
	return false
}

func (h History) HasPrefix(o History) bool {
	if len(o) > len(h) {
		return false
	}
	for i, e := range o {
		if e != h[i] {
			return false
		}
	}
	return true
}

func (h History) Equals(o History) bool {
	if len(h) != len(o) {
		return false
	}
	if h == nil || o == nil {
		return false
	}

	for i, e := range h {
		if e != o[i] {
			return false
		}
	}
	return true
}

func (h *History) Add(kind string, nv NameVersion) error {
	if h.Contains(nv) {
		return errors.ErrRecusion(kind, nv, *h)
	}
	*h = append(*h, nv)
	return nil
}

func (h History) Copy() History {
	return append(h[:0:0], h...)
}

func (h History) RemovePrefix(prefix History) History {
	for i, e := range prefix {
		if len(h) <= i || e != h[i] {
			return h[i:]
		}
	}
	return h[len(prefix):]
}

func (h History) Compare(o History) int {
	c, _ := h.Compare2(o)
	return c
}

func (h History) Compare2(o History) (int, bool) {
	for i, h := range h {
		if len(o) <= i {
			break
		}
		c := h.Compare(o[i])
		if c != 0 {
			return c, true
		}
	}
	return len(h) - len(o), false
}

////////////////////////////////////////////////////////////////////////////////

type HistoryElement interface {
	HistorySource
	GetKey() NameVersion
}

func SortHistoryElements(s interface{}) {
	rv := reflect.ValueOf(s)
	sort.Slice(s, func(i, j int) bool {
		return CompareHistoryElement(rv.Index(i).Interface(), rv.Index(j).Interface()) < 0
	})
}

func CompareHistorySource(a, b interface{}) int {
	aa := a.(HistorySource)
	ab := b.(HistorySource)

	return aa.GetHistory().Compare(ab.GetHistory())
}

func CompareHistoryElement(a, b interface{}) int {
	aa := a.(HistoryElement)
	ab := b.(HistoryElement)

	ha := aa.GetHistory()
	hb := ab.GetHistory()

	c, ok := ha.Compare2(hb)
	if ok {
		return c
	}
	k := 0
	switch {
	case c < 0:
		k = aa.GetKey().Compare(hb[len(ha)])
	case c > 0:
		k = ha[len(hb)].Compare(ab.GetKey())
	default:
		return aa.GetKey().Compare(ab.GetKey())
	}
	if k != 0 {
		return k
	}
	return c
}
