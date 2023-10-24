package packages_test

import (
	"fmt"
	"github.com/rstudio/python-distribution-parser/internal/packages"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSizeToString(t *testing.T) {
	tests := []struct {
		in  int64
		out string
	}{
		{
			0, "0.0 KB",
		},
		{
			512, "0.5 KB",
		},
		{
			1024, "1.0 KB",
		},
		{
			1024 * 1024, "1024.0 KB", // exactly 1024 KB
		},
		{
			1024*1024 + 1, "1.0 MB", // 1024 KB + 1
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%d is %s", test.in, test.out), func(t *testing.T) {
			in := test.in
			out := test.out
			t.Parallel()
			assert.EqualValues(t, out, packages.SizeToString(in))
		})
	}
}
