package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseSshConfigSuccess(t *testing.T) {
	testCases := []struct {
		name       string
		contents   string
		wantConfig []Host
	}{
		{
			name: "smallest possible",
			contents: `
Host foobar
  HostName 127.0.0.1
`,
			wantConfig: []Host{
				{
					"Host":     "foobar",
					"HostName": "127.0.0.1",
				},
			},
		},
		{
			name: "as generated by vagrant ssh-config",
			contents: `
Host foobar
  HostName 127.0.0.1
  User vagrant
  Port 2222
  UserKnownHostsFile /dev/null
  StrictHostKeyChecking no
  PasswordAuthentication no
  IdentityFile private_key
  IdentitiesOnly yes
`,
			wantConfig: []Host{
				{
					"Host":                   "foobar",
					"HostName":               "127.0.0.1",
					"User":                   "vagrant",
					"Port":                   "2222",
					"UserKnownHostsFile":     "/dev/null",
					"StrictHostKeyChecking":  "no",
					"PasswordAuthentication": "no",
					"IdentityFile":           "private_key",
					"IdentitiesOnly":         "yes",
				},
			},
		},
		{
			name: "two hosts",
			contents: `
Host foobar
  HostName 127.0.0.1

Host zoo
  HostName 1.2.3.4
`,
			wantConfig: []Host{
				{
					"Host":     "foobar",
					"HostName": "127.0.0.1",
				},
				{
					"Host":     "zoo",
					"HostName": "1.2.3.4",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rd := strings.NewReader(tc.contents)

			config, err := parseSshConfig(rd)

			if err != nil {
				t.Fatalf("error: have: %s; want: <no error>", err)
			}

			if diff := cmp.Diff(config, tc.wantConfig); diff != "" {
				t.Fatalf("\noutput mismatch (-have, +want)\n%s", diff)
			}
		})
	}
}

func TestParseSshConfigFailure(t *testing.T) {
	testCases := []struct {
		name     string
		contents string
		wantErr  string
	}{
		{
			name:     "empty file",
			contents: "",
			wantErr:  "parseSshConfig: empty file",
		},
		{
			name: "more than 2 tokens",
			contents: `
Host foobar
  HostName A B
`,
			wantErr: "parseSshConfig: line 'HostName A B': 3 tokens instead of 2",
		},
		{
			name:     "block must begin with Host",
			contents: "User vagrant",
			wantErr:  "parseSshConfig: line 'User vagrant': block must begin with 'Host'",
		},
		{
			name:     "empty Host block at beginning of file",
			contents: "Host A",
			wantErr:  "parseSshConfig: empty 'Host' block: map[Host:A]",
		},
		{
			name: "empty Host block in the middle of file",
			contents: `
Host A
  HostName B

Host C
`,
			wantErr: "parseSshConfig: empty 'Host' block: map[Host:C]",
		},
		{
			name: "duplicated key in same block",
			contents: `
Host A
  HostName B
  User vagrant
  HostName C
`,
			wantErr: "parseSshConfig: block 'Host A': duplicated k/v: 'HostName C', previous: 'HostName B'",
		},
		{
			name: "duplicated Host",
			contents: `
Host A
  HostName B

Host X
  HostName Y

Host A
  HostName C
`,
			wantErr: "parseSshConfig: duplicated block 'Host A'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rd := strings.NewReader(tc.contents)

			_, err := parseSshConfig(rd)

			have := "<no error>"
			if err != nil {
				have = err.Error()
			}
			if have != tc.wantErr {
				t.Fatalf("error: have: %s; want: %s", have, tc.wantErr)
			}
		})
	}
}

func TestHostGet(t *testing.T) {
	host := Host{
		"K": "V",
	}

	{
		have, err := host.Get("K")
		want := "V"
		if err != nil {
			t.Fatalf("get existing key: error: %s", err)
		}
		if have != want {
			t.Errorf("get existing key: have: %s; want: %s", have, want)
		}
	}

	{
		if _, err := host.Get("X"); err == nil {
			t.Fatal("get non-existing key: want: error; have: <no error>")
		}
	}
}

func TestHostGetDef(t *testing.T) {
	host := Host{
		"K": "V",
	}

	if have, want := host.GetDef("K", "default"), "V"; have != want {
		t.Errorf("get existing key: have: %s; want: %s", have, want)
	}

	if have, want := host.GetDef("X", "default"), "default"; have != want {
		t.Errorf("get non-existing key: have: %s; want: %s", have, want)
	}
}
