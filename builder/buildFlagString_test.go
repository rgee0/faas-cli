// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package builder

import "testing"

type inArgs struct {
	nocache    bool
	squash     bool
	httpProxy  string
	httpsProxy string
}

var buildFlagStringOptsTests = []struct {
	title       string
	inputArgs   inArgs
	expectedStr string
}{

	{
		title:       "No options added",
		inputArgs:   inArgs{false, false, "", ""},
		expectedStr: "",
	},
	{
		title:       "Https proxy option only",
		inputArgs:   inArgs{false, false, "", "192.168.0.1"},
		expectedStr: "--build-arg https_proxy=192.168.0.1 ",
	},
	{
		title:       "Http proxy option only",
		inputArgs:   inArgs{false, false, "192.168.0.1", ""},
		expectedStr: "--build-arg http_proxy=192.168.0.1 ",
	},
	{
		title:       "Both proxy types - nocache and squash are false",
		inputArgs:   inArgs{false, false, "192.168.0.1", "192.168.0.1"},
		expectedStr: "--build-arg http_proxy=192.168.0.1 --build-arg https_proxy=192.168.0.1 ",
	},
	{
		title:       "Both proxy types are empty - nocache is false and squash is true",
		inputArgs:   inArgs{false, true, "", ""},
		expectedStr: "--squash ",
	},
	{
		title:       "https_proxy with squash. http_proxy and nocache are empty and false",
		inputArgs:   inArgs{false, true, "", "192.168.0.1"},
		expectedStr: "--squash --build-arg https_proxy=192.168.0.1 ",
	},
	{
		title:       "http_proxy with squash. https_proxy and nocache are empty and false",
		inputArgs:   inArgs{false, true, "192.168.0.1", ""},
		expectedStr: "--squash --build-arg http_proxy=192.168.0.1 ",
	},
	{
		title:       "Both proxies with squash. nocache is false",
		inputArgs:   inArgs{false, true, "192.168.0.1", "192.168.0.1"},
		expectedStr: "--squash --build-arg http_proxy=192.168.0.1 --build-arg https_proxy=192.168.0.1 ",
	},
	{
		title:       "Both proxy types are empty - nocache is true and squash is false",
		inputArgs:   inArgs{true, false, "", ""},
		expectedStr: "--no-cache ",
	},
	{
		title:       "https_proxy with nocache. http_proxy and squash are empty and false",
		inputArgs:   inArgs{true, false, "", "192.168.0.1"},
		expectedStr: "--no-cache --build-arg https_proxy=192.168.0.1 ",
	},
	{
		title:       "http_proxy with nocache. https_proxy and squash are empty and false",
		inputArgs:   inArgs{true, false, "192.168.0.1", ""},
		expectedStr: "--no-cache --build-arg http_proxy=192.168.0.1 ",
	},
	{
		title:       "Both proxies with nocache. squash is false",
		inputArgs:   inArgs{true, false, "192.168.0.1", "192.168.0.1"},
		expectedStr: "--no-cache --build-arg http_proxy=192.168.0.1 --build-arg https_proxy=192.168.0.1 ",
	},
	{
		title:       "Squash and nocache are true.  Proxies omitted",
		inputArgs:   inArgs{true, true, "", ""},
		expectedStr: "--no-cache --squash ",
	},
	{
		title:       "Squash and nocache are true. Https_proxy option only",
		inputArgs:   inArgs{true, true, "", "192.168.0.1"},
		expectedStr: "--no-cache --squash --build-arg https_proxy=192.168.0.1 ",
	},
	{
		title:       "Squash and nocache are true. Http_proxy option only",
		inputArgs:   inArgs{true, true, "192.168.0.1", ""},
		expectedStr: "--no-cache --squash --build-arg http_proxy=192.168.0.1 ",
	},
	{
		title:       "Squash and nocache are true. Both proxies set",
		inputArgs:   inArgs{true, true, "192.168.0.1", "192.168.0.1"},
		expectedStr: "--no-cache --squash --build-arg http_proxy=192.168.0.1 --build-arg https_proxy=192.168.0.1 ",
	},
}

func TestBuildFlagString(t *testing.T) {
	for _, test := range buildFlagStringOptsTests {
		t.Run(test.title, func(t *testing.T) {

			flagString := buildFlagString(test.inputArgs.nocache, test.inputArgs.squash, test.inputArgs.httpProxy, test.inputArgs.httpsProxy)

			if flagString != test.expectedStr {
				t.Errorf("Flag String of (%t, %t, %s, %s) was incorrect, got: %s, want: %s.", test.inputArgs.nocache, test.inputArgs.squash, test.inputArgs.httpProxy, test.inputArgs.httpsProxy, flagString, test.expectedStr)
			}
		})
	}
}
