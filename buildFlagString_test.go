package main

import "testing"

func TestBuildFlagString(t *testing.T) {
	tables := []struct {
		nocache    bool
		squash     bool
		httpProxy  string
		httpsProxy string
		res        string
	}{
		{false, false, "", "", ""},
		{false, false, "", "192.168.0.1", "--build-arg https_proxy=192.168.0.1 "},
		{false, false, "192.168.0.1", "", "--build-arg http_proxy=192.168.0.1 "},
		{false, false, "192.168.0.1", "192.168.0.1", "--build-arg http_proxy=192.168.0.1 --build-arg https_proxy=192.168.0.1 "},
		{false, true, "", "", "--squash "},
		{false, true, "", "192.168.0.1", "--squash --build-arg https_proxy=192.168.0.1 "},
		{false, true, "192.168.0.1", "", "--squash --build-arg http_proxy=192.168.0.1 "},
		{false, true, "192.168.0.1", "192.168.0.1", "--squash --build-arg http_proxy=192.168.0.1 --build-arg https_proxy=192.168.0.1 "},
		{true, false, "", "", "--no-cache "},
		{true, false, "", "192.168.0.1", "--no-cache --build-arg https_proxy=192.168.0.1 "},
		{true, false, "192.168.0.1", "", "--no-cache --build-arg http_proxy=192.168.0.1 "},
		{true, false, "192.168.0.1", "192.168.0.1", "--no-cache --build-arg http_proxy=192.168.0.1 --build-arg https_proxy=192.168.0.1 "},
		{true, true, "", "", "--no-cache --squash "},
		{true, true, "", "192.168.0.1", "--no-cache --squash --build-arg https_proxy=192.168.0.1 "},
		{true, true, "192.168.0.1", "", "--no-cache --squash --build-arg http_proxy=192.168.0.1 "},
		{true, true, "192.168.0.1", "192.168.0.1", "--no-cache --squash --build-arg http_proxy=192.168.0.1 --build-arg https_proxy=192.168.0.1 "},
	}

	for _, table := range tables {
		total := buildFlagString(table.nocache, table.squash, table.httpProxy, table.httpsProxy)
		if total != table.res {
			t.Errorf("Flag String of (%t, %t, %s, %s) was incorrect, got: %s, want: %s.", table.nocache, table.squash, table.httpProxy, table.httpsProxy, total, table.res)
		}
	}
}
