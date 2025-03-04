package errgroup

import (
	"context"
	"errors"
	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func GetTestFunc(t *testing.T, name string, sleepTime time.Duration, ret error, HasPanic bool) func() error {
	return func() error {
		//t.Logf(" %v start", name)
		time.Sleep(sleepTime)
		if HasPanic {
			panic("panic occur")
		}
		//t.Logf(" %v end", name)
		return ret
	}
}
func TestErrgroup(t *testing.T) {
	type Field struct {
		name         string
		funcList     []func() error
		maxGoroutine int
		wantMaxTime  time.Duration
		wantMinTime  time.Duration
	}
	fields := []Field{
		{
			name: "normal",
			funcList: []func() error{
				GetTestFunc(t, "f-1s-nil", time.Second, nil, false),
				GetTestFunc(t, "f-2s-nil", time.Second*2, nil, false),
				GetTestFunc(t, "f-3s-nil", time.Second*3, nil, false),
			},
			wantMaxTime: time.Second*3 + time.Millisecond*50,
			wantMinTime: time.Second * 3,
		},
		{
			name: "one error",
			funcList: []func() error{
				GetTestFunc(t, "f-2s-nil", time.Second*2, nil, false),
				GetTestFunc(t, "f-1s-err", time.Second, errors.New("failed"), false),
				GetTestFunc(t, "f-300ms-nil", time.Millisecond*300, nil, false),
			},
			wantMaxTime: time.Second + time.Millisecond*50,
			wantMinTime: time.Second,
		},
		{
			name: "one panic",
			funcList: []func() error{
				GetTestFunc(t, "f-2s-nil", time.Second*2, nil, false),
				GetTestFunc(t, "f-1s-nil-panic", time.Second, nil, true),
				GetTestFunc(t, "f-300ms-nil", time.Millisecond*300, nil, false),
			},
			wantMaxTime: time.Second + time.Millisecond*50,
			wantMinTime: time.Second,
		},
		{
			name: "normal and max_goroutine=2",
			funcList: []func() error{
				GetTestFunc(t, "f-1s-nil", time.Second, nil, false),
				GetTestFunc(t, "f-1s-nil", time.Second, nil, false),
				GetTestFunc(t, "f-1s-nil", time.Second, nil, false),
			},
			maxGoroutine: 2,
			wantMaxTime:  time.Second*2 + time.Millisecond*50,
			wantMinTime:  time.Second * 2,
		},
		{
			name: "one error and max_goroutine=2",
			funcList: []func() error{
				GetTestFunc(t, "f-1s-nil", time.Second*2, nil, false),
				GetTestFunc(t, "f-1s-err", time.Second, nil, false),
				GetTestFunc(t, "f-1s-nil", time.Millisecond*300, errors.New("failed"), false),
			},
			maxGoroutine: 2,
			wantMaxTime:  time.Second*1 + time.Millisecond*350,
			wantMinTime:  time.Second*1 + time.Millisecond*300,
		},
	}
	for _, field := range fields {
		t.Run(field.name, func(t *testing.T) {
			PatchConvey(field.name, t, func() {
				now := time.Now()
				var interval time.Duration
				t.Logf("%v: start-time=%v", field.name, now)
				ctx := context.Background()
				g, ctx := WithContext(ctx)
				if field.maxGoroutine != 0 {
					g.SetLimit(field.maxGoroutine)
				}
				for _, f := range field.funcList {
					g.Go(f)
				}
				if wErr := g.Wait(); wErr != nil {
					t.Logf("%v err=%v", field.name, wErr)
				}
				interval = time.Now().Sub(now)
				So(interval, ShouldBeGreaterThanOrEqualTo, field.wantMinTime)
				So(interval, ShouldBeLessThan, field.wantMaxTime)
				t.Logf("%v: end-time=%v\ninterval=%v", field.name, time.Now(), interval)
			})
		})
	}
}
