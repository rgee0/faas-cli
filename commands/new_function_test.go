// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-cli/test"
)

const (
	SuccessMsg        = `(?m:Function created in folder)`
	InvalidYAMLMsg    = `is not valid YAML`
	InvalidYAMLMap    = `map is empty`
	IncludeUpperCase  = "function name can only contain a-z, 0-9 and dashes"
	NoFunctionName    = "please provide a name for the function"
	NoLanguage        = "you must supply a function language with the --lang flag"
	NoTemplates       = "no language templates were found. Please run 'faas-cli template pull'"
	InvalidFileSuffix = "when appending to a stack the suffix should be .yml or .yaml"
	InvalidFile       = "unable to find file: (.+)? - (.+)?"
	ListOptionOutput  = `Languages available as templates:
- dockerfile
- ruby`

	LangNotExistsOutput  = `(?m:is unavailable or not supported)`
	FunctionExistsOutput = `(Function (.+)? already exists in (.+)? file)`
)

type NewFunctionTest struct {
	title         string
	prefix        string
	funcName      string
	funcLang      string
	dirName       string
	expectedImage string
	expectedMsg   string
}

var NewFunctionTests = []NewFunctionTest{
	{
		title:         "new_1",
		funcName:      "new-test-1",
		funcLang:      "ruby",
		expectedImage: "new-test-1:latest",
		expectedMsg:   SuccessMsg,
	},
	{
		title:         "lowercase-dockerfile",
		funcName:      "lowercase-dockerfile",
		funcLang:      "dockerfile",
		expectedImage: "lowercase-dockerfile:latest",
		expectedMsg:   SuccessMsg,
	},
	{
		title:         "uppercase-dockerfile",
		funcName:      "uppercase-dockerfile",
		funcLang:      "dockerfile",
		expectedImage: "uppercase-dockerfile:latest",
		expectedMsg:   SuccessMsg,
	},
	{
		title:         "func-with-prefix",
		funcName:      "func-with-prefix",
		prefix:        " username ",
		funcLang:      "dockerfile",
		expectedImage: "username/func-with-prefix:latest",
		expectedMsg:   SuccessMsg,
	},
	{
		title:         "func-with-whitespace-only-prefix",
		funcName:      "func-with-whitespace-only-prefix",
		prefix:        " ",
		funcLang:      "dockerfile",
		expectedImage: "func-with-whitespace-only-prefix:latest",
		expectedMsg:   SuccessMsg,
	},
	{
		title:         "long-name-with-hyphens",
		funcName:      "long-name-with-hyphens",
		dirName:       "customname",
		prefix:        " ",
		funcLang:      "dockerfile",
		expectedImage: "long-name-with-hyphens:latest",
		expectedMsg:   SuccessMsg,
	},
	{
		title:       "invalid_1",
		funcName:    "new-test-invalid-1",
		funcLang:    "dockerfilee",
		expectedMsg: LangNotExistsOutput,
	},
	{
		title:       "test_Uppercase",
		funcName:    "test_Uppercase",
		funcLang:    "dockerfile",
		expectedMsg: IncludeUpperCase,
	},
	{
		title:       "no-function-name",
		funcName:    "",
		funcLang:    "",
		expectedMsg: NoFunctionName,
	},
	{
		title:       "no-language",
		funcName:    "no-language",
		funcLang:    "",
		expectedMsg: NoLanguage,
	},
}

func runNewFunctionTest(t *testing.T, nft NewFunctionTest) {
	funcName := nft.funcName
	funcLang := nft.funcLang
	dirName := nft.dirName
	imagePrefix := nft.prefix
	var funcYAML string
	funcYAML = funcName + ".yml"

	cmdParameters := []string{
		"new",
		"--lang=" + funcLang,
		"--gateway=" + defaultGateway,
		"--prefix=" + imagePrefix,
	}
	if len(dirName) != 0 {
		cmdParameters = append(cmdParameters, "--handler="+dirName)
	} else {
		dirName = funcName
	}
	if len(funcName) != 0 {
		cmdParameters = append(cmdParameters, funcName)
	}

	faasCmd.SetArgs(cmdParameters)
	execErr := faasCmd.Execute()

	if nft.expectedMsg == SuccessMsg {

		// Make sure that the folder and file was created:
		if _, err := os.Stat("./" + dirName); os.IsNotExist(err) {
			t.Fatalf("%s/ directory was not created", dirName)
		}

		// Check that the Dockerfile was created
		if funcLang == "Dockerfile" || funcLang == "dockerfile" {
			if _, err := os.Stat("./" + dirName + "/Dockerfile"); os.IsNotExist(err) {
				t.Fatalf("Dockerfile language should create a Dockerfile for you: %s", funcName)
			}
		}

		if _, err := os.Stat(funcYAML); os.IsNotExist(err) {
			t.Fatalf("\"%s\" yaml file was not created", funcYAML)
		}

		// Make sure that the information in the YAML file is correct:
		parsedServices, err := stack.ParseYAMLFile(funcYAML, "", "")
		if err != nil {
			t.Fatalf("Couldn't open modified YAML file \"%s\" due to error: %v", funcYAML, err)
		}
		services := *parsedServices

		var testServices stack.Services
		testServices.Provider = stack.Provider{Name: "faas", GatewayURL: defaultGateway}
		if !reflect.DeepEqual(services.Provider, testServices.Provider) {
			t.Fatalf("YAML `provider` section was not created correctly for file %s: got %v", funcYAML, services.Provider)
		}

		testServices.Functions = make(map[string]stack.Function)
		testServices.Functions[funcName] = stack.Function{Language: funcLang, Image: nft.expectedImage, Handler: "./" + dirName}
		if !reflect.DeepEqual(services.Functions[funcName], testServices.Functions[funcName]) {
			t.Fatalf("YAML `functions` section was not created correctly for file %s, got %v", funcYAML, services.Functions[funcName])
		}
	} else {
		// Validate new function output
		if found, err := regexp.MatchString(nft.expectedMsg, execErr.Error()); err != nil || !found {
			t.Fatalf("Output is not as expected: %s\n", execErr)
		}
	}

}

func Test_newFunctionTests(t *testing.T) {
	// Download templates
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	for _, testcase := range NewFunctionTests {
		t.Run(testcase.title, func(t *testing.T) {
			defer tearDownNewFunction(t, testcase.funcName)
			runNewFunctionTest(t, testcase)
		})
	}
}

func Test_newFunctionListCmds(t *testing.T) {
	// Download templates
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	cmdParameters := []string{
		"new",
		"--list",
	}

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs(cmdParameters)
		faasCmd.Execute()
	})

	// Validate command output
	if !strings.HasPrefix(stdOut, ListOptionOutput) {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func Test_newFunctionListNoTemplates(t *testing.T) {
	cmdParameters := []string{
		"new",
		"--list",
	}

	faasCmd.SetArgs(cmdParameters)
	stdOut := faasCmd.Execute().Error()

	// Validate command output
	if !strings.HasPrefix(stdOut, NoTemplates) {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func Test_languageNotExists(t *testing.T) {
	// Download templates
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	// Attempt to create a function with a non-existing language
	cmdParameters := []string{
		"new",
		"samplename",
		"--lang=bash",
		"--gateway=" + defaultGateway,
		"--list=false",
	}

	faasCmd.SetArgs(cmdParameters)
	stdOut := faasCmd.Execute().Error()

	// Validate new function output
	if found, err := regexp.MatchString(LangNotExistsOutput, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func Test_appendInvalidSuffix(t *testing.T) {
	const functionName = "samplefunc"
	const functionLang = "ruby"

	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	// Create function
	parameters := []string{
		"new",
		functionName,
		"--lang=" + functionLang,
		"--append=" + functionName + ".txt",
	}
	faasCmd.SetArgs(parameters)
	stdOut := faasCmd.Execute().Error()

	if found, err := regexp.MatchString(InvalidFileSuffix, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func Test_appendInvalidFile(t *testing.T) {
	const functionName = "samplefunc"
	const functionLang = "ruby"

	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	// Create function
	parameters := []string{
		"new",
		functionName,
		"--lang=" + functionLang,
		"--append=" + functionLang + ".yml",
	}
	faasCmd.SetArgs(parameters)
	stdOut := faasCmd.Execute().Error()

	if found, err := regexp.MatchString(InvalidFile, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func Test_templateParsingyYAMLEnvs(t *testing.T) {

	var services *stack.Services
	var tpl bytes.Buffer

	newFunction := stack.Function{
		Name:     "functionName",
		Handler:  "./handlerDir",
		Language: "language",
		Image:    "imageName",
		Environment: map[string]string{
			"alex": "ellis",
		},
		EnvironmentFile: []string{
			"alex.yml",
			"ellis.yml",
		},
		Secrets: []string{
			"alex",
			"ellis",
		},
		Constraints: &[]string{
			"node.platform.os == linux",
		},
		Labels: &map[string]string{
			"alex": "ellis",
		},
		Limits: &stack.FunctionResources{
			Memory: "40Mi",
			CPU:    "100m",
		},
		Requests: &stack.FunctionResources{
			Memory: "30Mi",
			CPU:    "200m",
		},
		Annotations: &map[string]string{
			"topic":  "kafka.payments-received",
			"people": "burt_and_ernie",
		},
		ReadOnlyRootFilesystem: false,
		SkipBuild:              false,
	}

	services = &stack.Services{
		Provider: stack.Provider{
			Name:       "faas",
			GatewayURL: gateway,
		},
		Functions: make(map[string]stack.Function),
	}

	services.Functions["functionName"] = newFunction

	tmplt := template.Must(template.New("stack").Parse(stackTmpl))
	err := tmplt.Execute(&tpl, services)

	if err != nil {
		t.Fatalf("Executing the template failed: %s\n", err)
	}

	yamlStr := tpl.String()
	if !(strings.Contains(yamlStr, "environment")) || !(strings.Contains(yamlStr, "alex")) || !(strings.Contains(yamlStr, "ellis")) {
		t.Fatalf("Missing elements in the yaml, expected all of: '%s', '%s' & '%s' \n Found:\n %s \n", "environment", "alex", "ellis", yamlStr)
	}
	fmt.Printf("%s", yamlStr)

}
func Test_duplicateFunctionName(t *testing.T) {
	resetForTest()

	const functionName = "samplefunc"
	const functionLang = "ruby"

	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)
	defer tearDownNewFunction(t, functionName)

	// Create function
	parameters := []string{
		"new",
		functionName,
		"--lang=" + functionLang,
	}
	faasCmd.SetArgs(parameters)
	faasCmd.Execute()

	// Attempt to create duplicate function
	parameters = append(parameters, "--append="+functionName+".yml")
	faasCmd.SetArgs(parameters)
	stdOut := faasCmd.Execute().Error()

	if found, err := regexp.MatchString(FunctionExistsOutput, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func tearDownNewFunction(t *testing.T, functionName string) {
	if _, err := os.Stat(".gitignore"); err == nil {
		if err := os.Remove(".gitignore"); err != nil {
			t.Log(err)
		}
	}
	hDir := handlerDir
	if len(hDir) == 0 {
		hDir = functionName
	}
	if _, err := os.Stat(hDir); err == nil {
		if err := os.RemoveAll(hDir); err != nil {
			t.Log(err)
		}
	}
	functionYaml := functionName + ".yml"
	if _, err := os.Stat(functionYaml); err == nil {
		if err := os.Remove(functionYaml); err != nil {
			t.Log(err)
		}
	}
	handlerDir = ""
}
