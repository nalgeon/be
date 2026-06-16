package be_test

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/nalgeon/be"
)

// mockTB is a mock implementation of testing.TB
// to capture test failures.
type mockTB struct {
	testing.TB
	failed bool
	fatal  bool
	msg    string
}

func (m *mockTB) Helper() {}

func (m *mockTB) Fatal(args ...any) {
	m.fatal = true
	m.Error(args...)
}

func (m *mockTB) Fatalf(format string, args ...any) {
	m.fatal = true
	m.Errorf(format, args...)
}

func (m *mockTB) Error(args ...any) {
	m.failed = true
	m.msg = fmt.Sprint(args...)
}

func (m *mockTB) Errorf(format string, args ...any) {
	m.failed = true
	m.msg = fmt.Sprintf(format, args...)
}

// intType wraps an int value.
type intType struct {
	val int
}

// noisy provides an Equal method.
type noisy struct {
	val   int
	noise float64
}

func newNoisy(val int) noisy {
	return noisy{val: val, noise: rand.Float64()}
}

func (n noisy) Equal(other noisy) bool {
	return n.val == other.val
}

// errType is a custom error type.
type errType string

func (e errType) Error() string {
	return string(e)
}

func TestEqual(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		now := time.Now()
		val := 42

		testCases := map[string]struct {
			got  any
			want any
		}{
			"integer":     {got: 42, want: 42},
			"string":      {got: "hello", want: "hello"},
			"bool":        {got: true, want: true},
			"struct":      {got: intType{42}, want: intType{42}},
			"pointer":     {got: &val, want: &val},
			"nil slice":   {got: []int(nil), want: []int(nil)},
			"byte slice":  {got: []byte("abc"), want: []byte("abc")},
			"int slice":   {got: []int{42, 84}, want: []int{42, 84}},
			"time.Time":   {got: now, want: now},
			"nil":         {got: nil, want: nil},
			"nil pointer": {got: (*int)(nil), want: (*int)(nil)},
			"nil map":     {got: map[string]int(nil), want: map[string]int(nil)},
			"nil chan":    {got: (chan int)(nil), want: (chan int)(nil)},
			"empty map":   {got: map[string]int{}, want: map[string]int{}},
			"map":         {got: map[string]int{"a": 42}, want: map[string]int{"a": 42}},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				tb := &mockTB{}
				be.Equal(tb, tc.got, tc.want)
				if tb.failed {
					t.Errorf("%#v vs %#v: should have passed", tc.got, tc.want)
				}
			})
		}
	})
	t.Run("non-equal", func(t *testing.T) {
		val1, val2 := 42, 84
		now := time.Now()

		testCases := map[string]struct {
			got  any
			want any
			msg  string
		}{
			"integer": {
				got: 42, want: 84,
				msg: "got: 42; want: 84",
			},
			"int32 vs int64": {
				got: int32(42), want: int64(42),
				msg: "got: 42; want: 42",
			},
			"int vs string": {
				got: 42, want: "42",
				msg: `got: 42; want: "42"`,
			},
			"string": {
				got: "hello", want: "world",
				msg: `got: "hello"; want: "world"`,
			},
			"bool": {
				got: true, want: false,
				msg: "got: true; want: false",
			},
			"struct": {
				got: intType{42}, want: intType{84},
				msg: "got: be_test.intType{val:42}; want: be_test.intType{val:84}",
			},
			"pointer": {
				got: &val1, want: &val2,
			},
			"byte slice": {
				got: []byte("abc"), want: []byte("abd"),
				msg: `got: []byte{0x61, 0x62, 0x63}; want: []byte{0x61, 0x62, 0x64}`,
			},
			"int slice": {
				got: []int{42, 84}, want: []int{84, 42},
				msg: `got: []int{42, 84}; want: []int{84, 42}`,
			},
			"int slice vs any slice": {
				got: []int{42, 84}, want: []any{42, 84},
				msg: `got: []int{42, 84}; want: []interface {}{42, 84}`,
			},
			"time.Time": {
				got: now, want: now.Add(time.Second),
			},
			"nil vs non-nil": {
				got: nil, want: 42,
				msg: "got: <nil>; want: 42",
			},
			"non-nil vs nil": {
				got: 42, want: nil,
				msg: "got: 42; want: <nil>",
			},
			"nil vs empty": {
				got: []int(nil), want: []int{},
				msg: "got: []int(nil); want: []int{}",
			},
			"map": {
				got: map[string]int{"a": 42}, want: map[string]int{"a": 84},
				msg: `got: map[string]int{"a":42}; want: map[string]int{"a":84}`,
			},
			"chan": {
				got: make(chan int), want: make(chan int),
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				tb := &mockTB{}
				be.Equal(tb, tc.got, tc.want)
				if !tb.failed {
					t.Errorf("%#v vs %#v: should have failed", tc.got, tc.want)
				}
				if tb.fatal {
					t.Error("should not be fatal")
				}
				if tc.msg != "" && tb.msg != tc.msg {
					t.Errorf("got: %q; want: %q", tb.msg, tc.msg)
				}
			})
		}
	})
	t.Run("time", func(t *testing.T) {
		// date1 and date2 represent the same point in time,
		date1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2025, 1, 1, 5, 0, 0, 0, time.FixedZone("UTC+5", 5*3600))
		tb := &mockTB{}
		be.Equal(tb, date1, date2)
		if tb.failed {
			t.Errorf("%#v vs %#v: should have passed", date1, date2)
		}
	})
	t.Run("equaler", func(t *testing.T) {
		t.Run("equal", func(t *testing.T) {
			tb := &mockTB{}
			n1, n2 := newNoisy(42), newNoisy(42)
			be.Equal(tb, n1, n2)
			if tb.failed {
				t.Errorf("%#v vs %#v: should have passed", n1, n2)
			}
		})
		t.Run("non-equal", func(t *testing.T) {
			tb := &mockTB{}
			n1, n2 := newNoisy(42), newNoisy(84)
			be.Equal(tb, n1, n2)
			if !tb.failed {
				t.Errorf("%#v vs %#v: should have failed", n1, n2)
			}
			if tb.fatal {
				t.Error("should not be fatal")
			}
		})
	})
	t.Run("no wants", func(t *testing.T) {
		tb := &mockTB{}
		be.Equal(tb, 42)
		if !tb.failed {
			t.Error("should have failed")
		}
		if !tb.fatal {
			t.Error("should be fatal")
		}
		wantMsg := "no wants given"
		if tb.msg != wantMsg {
			t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
		}
	})
	t.Run("multiple wants", func(t *testing.T) {
		t.Run("all equal", func(t *testing.T) {
			tb := &mockTB{}
			x := 2 * 3 * 7
			be.Equal(tb, x, 42, 42, 42)
			if tb.failed {
				t.Error("should have passed")
			}
		})
		t.Run("some equal", func(t *testing.T) {
			tb := &mockTB{}
			x := 2 * 3 * 7
			be.Equal(tb, x, 21, 42, 84)
			if tb.failed {
				t.Error("should have passed")
			}
		})
		t.Run("none equal", func(t *testing.T) {
			tb := &mockTB{}
			x := 2 * 3 * 7
			be.Equal(tb, x, 11, 12, 13)
			if !tb.failed {
				t.Error("should have failed")
			}
			if tb.fatal {
				t.Error("should not be fatal")
			}
			wantMsg := "got: 42; want any of: [11 12 13]"
			if tb.msg != wantMsg {
				t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
			}
		})
	})
}

func TestErr(t *testing.T) {
	t.Run("want nil", func(t *testing.T) {
		t.Run("got nil", func(t *testing.T) {
			tb := &mockTB{}
			be.Err(tb, nil, nil)
			if tb.failed {
				t.Errorf("failed: %s", tb.msg)
			}
		})
		t.Run("got error", func(t *testing.T) {
			tb := &mockTB{}
			err := errors.New("oops")
			be.Err(tb, err, nil)
			if !tb.failed {
				t.Error("should have failed")
			}
			if !tb.fatal {
				t.Error("should be fatal")
			}
			wantMsg := "unexpected error: oops"
			if tb.msg != wantMsg {
				t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
			}
		})
	})
	t.Run("want error", func(t *testing.T) {
		t.Run("got nil", func(t *testing.T) {
			tb := &mockTB{}
			err := errors.New("oops")
			be.Err(tb, nil, err)
			if !tb.failed {
				t.Error("should have failed")
			}
			if tb.fatal {
				t.Error("should not be fatal")
			}
			wantMsg := `got: <nil>; want: error`
			if tb.msg != wantMsg {
				t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
			}
		})
		t.Run("same error", func(t *testing.T) {
			tb := &mockTB{}
			err := errors.New("oops")
			be.Err(tb, err, err)
			if tb.failed {
				t.Errorf("failed: %s", tb.msg)
			}
		})
		t.Run("wrapped error", func(t *testing.T) {
			tb := &mockTB{}
			err := errors.New("oops")
			wrappedErr := fmt.Errorf("wrapped: %w", err)
			be.Err(tb, wrappedErr, err)
			if tb.failed {
				t.Errorf("failed: %s", tb.msg)
			}
		})
		t.Run("different value", func(t *testing.T) {
			tb := &mockTB{}
			err1 := errors.New("error 1")
			err2 := errors.New("error 2")
			be.Err(tb, err1, err2)
			if !tb.failed {
				t.Error("should have failed")
			}
			if tb.fatal {
				t.Error("should not be fatal")
			}
			wantMsg := "got: *errors.errorString(error 1); want: *errors.errorString(error 2)"
			if tb.msg != wantMsg {
				t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
			}
		})
		t.Run("different type", func(t *testing.T) {
			tb := &mockTB{}
			err1 := errors.New("oops")
			err2 := errType("oops")
			be.Err(tb, err1, err2)
			if !tb.failed {
				t.Error("should have failed")
			}
			if tb.fatal {
				t.Error("should not be fatal")
			}
			wantMsg := "got: *errors.errorString(oops); want: be_test.errType(oops)"
			if tb.msg != wantMsg {
				t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
			}
		})
	})
	t.Run("want string", func(t *testing.T) {
		t.Run("contains", func(t *testing.T) {
			tb := &mockTB{}
			err := errors.New("the night is dark")
			be.Err(tb, err, "night is")
			if tb.failed {
				t.Errorf("failed: %s", tb.msg)
			}
		})
		t.Run("does not contain", func(t *testing.T) {
			tb := &mockTB{}
			err := errors.New("the night is dark")
			be.Err(tb, err, "day")
			if !tb.failed {
				t.Error("should have failed")
			}
			if tb.fatal {
				t.Error("should not be fatal")
			}
			wantMsg := `got: "the night is dark"; want: "day"`
			if tb.msg != wantMsg {
				t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
			}
		})
	})
	t.Run("want type", func(t *testing.T) {
		t.Run("same type", func(t *testing.T) {
			tb := &mockTB{}
			err := errType("oops")
			be.Err(tb, err, reflect.TypeFor[errType]())
			if tb.failed {
				t.Errorf("failed: %s", tb.msg)
			}
		})
		t.Run("different type", func(t *testing.T) {
			tb := &mockTB{}
			err := errType("oops")
			be.Err(tb, err, reflect.TypeFor[*fs.PathError]())
			if !tb.failed {
				t.Error("should have failed")
			}
			if tb.fatal {
				t.Error("should not be fatal")
			}
			wantMsg := "got: be_test.errType; want: *fs.PathError"
			if tb.msg != wantMsg {
				t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
			}
		})
	})
	t.Run("unsupported want", func(t *testing.T) {
		tb := &mockTB{}
		var want int
		be.Err(tb, errors.New("oops"), want)
		if !tb.failed {
			t.Error("should have failed")
		}
		if tb.fatal {
			t.Error("should not be fatal")
		}
		wantMsg := "unsupported want type: int"
		if tb.msg != wantMsg {
			t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
		}
	})
	t.Run("no wants", func(t *testing.T) {
		t.Run("got error", func(t *testing.T) {
			tb := &mockTB{}
			err := errors.New("oops")
			be.Err(tb, err)
			if tb.failed {
				t.Error("should have passed")
			}

		})
		t.Run("got nil", func(t *testing.T) {
			tb := &mockTB{}
			var err error
			be.Err(tb, err)
			if !tb.failed {
				t.Error("should have failed")
			}
			if tb.fatal {
				t.Error("should not be fatal")
			}
			wantMsg := "got: <nil>; want: error"
			if tb.msg != wantMsg {
				t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
			}
		})
	})
	t.Run("multiple wants", func(t *testing.T) {
		t.Run("all match", func(t *testing.T) {
			tb := &mockTB{}
			err := errType("oops")
			be.Err(tb, err, errType("oops"), "oops", reflect.TypeFor[errType]())
			if tb.failed {
				t.Error("should have passed")
			}
		})
		t.Run("some match", func(t *testing.T) {
			tb := &mockTB{}
			err := errType("oops")
			be.Err(tb, err, errType("oops"), 42, reflect.TypeFor[errType]())
			if tb.failed {
				t.Error("should have passed")
			}
		})
		t.Run("none match", func(t *testing.T) {
			tb := &mockTB{}
			err := errType("oops")
			be.Err(tb, err, errType("failed"), 42, reflect.TypeFor[*fs.PathError]())
			if !tb.failed {
				t.Error("should have failed")
			}
			if tb.fatal {
				t.Error("should not be fatal")
			}
			wantMsg := "got: be_test.errType(oops); want any of: [failed 42 *fs.PathError]"
			if tb.msg != wantMsg {
				t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
			}
		})
	})
}

func TestTrue(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		tb := &mockTB{}
		be.True(tb, true)
		if tb.failed {
			t.Errorf("failed: %s", tb.msg)
		}
	})
	t.Run("false", func(t *testing.T) {
		tb := &mockTB{}
		be.True(tb, false)
		if !tb.failed {
			t.Error("should have failed")
		}
		if tb.fatal {
			t.Error("should not be fatal")
		}
		wantMsg := "got: false; want: true"
		if tb.msg != wantMsg {
			t.Errorf("got: %q; want: %q", tb.msg, wantMsg)
		}
	})
	t.Run("expression", func(t *testing.T) {
		tb := &mockTB{}
		f := func() int { return 42 }
		be.True(tb, (f() == 42))
		if tb.failed {
			t.Errorf("failed: %s", tb.msg)
		}
	})
}
