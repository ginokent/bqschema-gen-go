package main

import (
	"cloud.google.com/go/bigquery"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const (
	optNameProjectID                    = "project"
	optNameDataset                      = "dataset"
	optNameKeyFile                      = "keyfile"
	optNameOutputPath                   = "output"
	envNameGoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
)

var (
	vOptProjectID  string
	vOptDataset    string
	vOptKeyFile    string
	vOptOutputPath string
)

func init() {
	flag.StringVar(&vOptProjectID, optNameProjectID, "", "")
	flag.StringVar(&vOptDataset, optNameDataset, "", "")
	flag.StringVar(&vOptKeyFile, optNameKeyFile, "", "path to service account json key file")
	flag.StringVar(&vOptOutputPath, optNameOutputPath, "", "path to output the generated code")
	flag.Parse()
}

const (
	goFileContentHeader = `// Code generated by bqtableschema.go; DO NOT EDIT.

package bqtableschema
`
)

func main() {
	mainCtx := context.Background()

	if err := run(mainCtx); err != nil {
		log.Fatalf("%s: %w\n", FuncNameWithFileInfo(), err)
	}
}

// run is effectively a `main` function.
// It is separated from the `main` function because of addressing an issue where` defer` is not executed when `os.Exit` is executed.
func run(mainCtx context.Context) error {
	// NOTE(djeeno): output 1
	fmt.Printf("%s\n", goFileContentHeader)

	// NOTE(djeeno): if passed -keyfile,
	if vOptKeyFile != "" {
		if err := os.Setenv(envNameGoogleApplicationCredentials, vOptKeyFile); err != nil {
			log.Fatalf("%s: %w\n", FuncNameWithFileInfo(), err)
		}
	}

	envKeyFile := os.Getenv(envNameGoogleApplicationCredentials)

	if envKeyFile == "" {
		log.Fatalf("%s: set environment variable %s, or set option -%s", FuncNameWithFileInfo(), envNameGoogleApplicationCredentials, optNameKeyFile)
	}

	cred, err := newGoogleApplicationCredentials(envKeyFile)
	if err != nil {
		log.Fatalf("%s: %w\n", FuncNameWithFileInfo(), err)
	}

	var projectID string
	if vOptProjectID != "" {
		projectID = vOptProjectID
	} else {
		projectID = cred.ProjectID
	}

	c, err := bigquery.NewClient(mainCtx, projectID)
	defer func() {
		if err := c.Close(); err != nil {
			log.Fatalf("%s: %v\n", FuncNameWithFileInfo(), err)
		}
	}()
	if err != nil {
		log.Fatalf("%s: %v\n\n", FuncNameWithFileInfo(), err)
	}

	_ = c

	return nil
}

type googleApplicationCredentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

func newGoogleApplicationCredentials(path string) (*googleApplicationCredentials, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", FuncNameWithFileInfo(), err)
	}

	bytea, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", FuncNameWithFileInfo(), err)
	}

	cred := googleApplicationCredentials{}
	if err := json.Unmarshal(bytea, &cred); err != nil {
		return nil, fmt.Errorf("getGoogleProject: %w", err)
	}

	return &cred, nil
}

// ResolveEnvs resolves environment variables from the arguments passed as environment variable names.
func ResolveEnvs(keys ...string) (map[string]string, error) {
	envs := map[string]string{}

	for _, key := range keys {
		envs[key] = os.Getenv(key)
		if envs[key] == "" {
			return nil, fmt.Errorf("%s: environment variable %s is empty", FuncNameWithFileInfo(), key)
		}
	}

	return envs, nil
}

// MergeMap merge map[string]string
func MergeMap(sideToBeMerged, sideToMerge map[string]string) map[string]string {
	m := map[string]string{}

	for k, v := range sideToBeMerged {
		m[k] = v
	}
	for k, v := range sideToMerge {
		m[k] = v
	}
	return (m)
}

func FuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	return runtime.FuncForPC(pc).Name()
}

func FuncNameWithFileInfo() string {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return fmt.Sprintf("%s[%s:%d]", "null", "null", 0)
	}
	return fmt.Sprintf("%s[%s:%d]", runtime.FuncForPC(pc).Name(), filepath.Base(file), line)
}
