package parse_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"s1mpleasia.com/tinydb/parse"
)

func TestPredParser(t *testing.T) {
	// t.Parallel()

	for _, tt := range []struct {
		input     string
		wantError bool
	}{
		{
			input:     "age = 28",
			wantError: false,
		},

		{
			input:     "age = 20 AND name='Alice'",
			wantError: false,
		},

		{
			input:     "age = 20 AND name='Alice' and 1=2",
			wantError: false,
		},

		{
			input:     "is_expired",
			wantError: true,
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			// t.Parallel()

			p, err := parse.NewPredParser(tt.input)
			require.NoError(t, err)

			err = p.Predicate()

			if tt.wantError {
				fmt.Println(err)
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
