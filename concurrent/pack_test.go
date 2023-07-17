package concurrent

import (
	"context"
	"strings"
)

type Pack1 struct{}

func (this *Pack1) Pack(ctx context.Context, param, taskContext interface{}) error {
	t := taskContext.(*TaskContext)
	t.res1 = strings.Join([]string{
		t.a,
		t.b,
		t.c,
		t.d,
		t.e,
		t.f,
		t.g,
		t.h,
	}, ":")

	return nil
}

type Pack2 struct{}

func (this *Pack2) Pack(ctx context.Context, param, taskContext interface{}) error {
	t := taskContext.(*TaskContext)
	if t.res1 == "" {
		t.res2 = "shit"
	} else {
		t.res2 = "done"
	}

	return nil
}

type Pack3 struct{}

func (this *Pack3) Pack(ctx context.Context, param, taskContext interface{}) error {
	t := taskContext.(*TaskContext)
	if t.res2 == "shit" || t.res2 == "" {
		t.res3 = "shit"
	} else {
		t.res3 = "done"
	}

	return nil
}
