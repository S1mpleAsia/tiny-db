package parse_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"s1mpleasia.com/tinydb/parse"
)

func TestParserQuery(t *testing.T) {
	// t.Parallel()

	for _, tt := range []struct {
		input     string
		wantQuery string
		wantError bool
	}{
		{
			input:     "SELECT sid, SName, age FROM STUDENT",
			wantQuery: "select sid, sname, age from student",
			wantError: false,
		},

		{
			input:     "SELECT sname from STUDENT where age = 20",
			wantQuery: "select sname from student where age = 20",
			wantError: false,
		},

		{
			input:     "SELECT sname FROM student WHERE age = 20 AND did = 3",
			wantQuery: "select sname from student where age = 20 and did = 3",
			wantError: false,
		},

		{
			input:     "select sid, sname, did, dname FROM student, dept WHERE sname = 'John'",
			wantQuery: "select sid, sname, did, dname from student, dept where sname = 'John'",
			wantError: false,
		},

		{
			input:     "select sid, sname, did, dname, FROM student, dept WHERE sname = 'John'",
			wantQuery: "select sid, sname, did, dname from student, dept where sname = 'John'",
			wantError: false,
		},

		{
			input:     "SELECT * FROM STUDENT",
			wantError: true,
		},

		{
			input:     "SELECT sid,, FROM STUDENT",
			wantError: true,
		},

		{
			input:     "SELECT sid STUDENT",
			wantError: true,
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			// t.Parallel()

			p, err := parse.NewParser(tt.input)
			require.NoError(t, err)

			query, err := p.Query()

			if tt.wantError {
				var errBadSyntax *parse.BadSyntaxError
				assert.ErrorAs(t, err, &errBadSyntax, fmt.Errorf("[%s] test error: [%w]", tt.input, err))
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantQuery, query.String())
			}
		})
	}
}
