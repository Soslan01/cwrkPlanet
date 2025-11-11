package tests

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/cwrk-planet/logger/pkg/logger"
)

func captureStdOut(fn func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = orig
	}()

	fn()

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func toAttrsFromCtx(ctx context.Context) []any {
	attrs := logger.AttrsFromCtx(ctx)
	result := make([]any, len(attrs))
	for i, attr := range attrs {
		result[i] = attr
	}

	return result
}
