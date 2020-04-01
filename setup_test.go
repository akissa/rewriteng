package rewriteng

import (
	"testing"

	"github.com/caddyserver/caddy"
)

func TestSetup(t *testing.T) {
}

func TestParse(t *testing.T) {
	tests := []struct {
		inputFileRules string
		shouldErr      bool
	}{
		// parse errors
		{`rewriteng`, true},
		{`rewriteng X`, true},
		{`rewriteng ANY`, true},
		{`rewriteng ANY X`, true},
		{`rewriteng ANY ANY xxx`, true},
		{
			`rewriteng ANY ANY regex example\.org {
				answer regex example\.com example.org
			}`,
			true,
		},
		{
			`rewriteng ANY ANY regexx example\.org example.com {
				answer regex example\.com example.org
			}`,
			true,
		},
		{
			`rewriteng ANY ANY regex example\.org example.com {
				answer regexx example\.com example.org
			}`,
			true,
		},
		{
			`rewriteng ANY ANY regex example\.org example.com {
			}`,
			true,
		},
		// pass
		{
			`rewriteng ANY ANY regex example\.org example.com {
				answer regex example\.com example.org
				additional regex (.*)\.example\.com {1}.example.org
				additional regex (.*)\.example\.net {1}.example.org
				authority regex (.*)\.example\.net {1}.example.org
			}`,
			false,
		},
		// pass
		{
			`rewriteng ANY ANY regex example\.org example.com {
						answer regex example\.com example.org
					}`,
			false,
		},
		// pass
		{
			`rewriteng ANY ANY noop example.org example.org {
				answer noop example\.com example.org
			}`,
			false,
		},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.inputFileRules)
		_, err := parse(c)
		if err == nil && test.shouldErr {
			t.Fatalf("Test %d expected errors, but got no error\n---\n%s\n", i, test.inputFileRules)
		} else if err != nil && !test.shouldErr {
			t.Fatalf("Test %d expected no errors, but got '%v'", i, err)
		}
	}
}
