//go:generate go run github.com/ginokent/bqschema-gen-go

package main

import (
	"context"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"golang.org/x/tools/imports"
	"google.golang.org/api/iterator"
)

const (
	// optName
	optNameProjectID  = "project"
	optNameDataset    = "dataset"
	optNameOutputFile = "output"
	optNameDebug      = "debug"
	// envName
	envNameGCloudProjectID = "GCLOUD_PROJECT_ID"
	envNameBigQueryDataset = "BIGQUERY_DATASET"
	envNameOutputFile      = "OUTPUT_FILE"
	envNameDebug           = "DEBUG"
	// defaultValue
	defaultValueEmpty      = ""
	defaultValueOutputFile = "bqschema.generated.go"
	defaultValueDebug      = "false"
)

var (
	// optValue
	optValueProjectID  = flag.String(optNameProjectID, defaultValueEmpty, "")
	optValueDataset    = flag.String(optNameDataset, defaultValueEmpty, "")
	optValueOutputPath = flag.String(optNameOutputFile, defaultValueEmpty, "path to output the generated code")
)

func main() {

	ctx := context.Background()

	if err := Run(ctx); err != nil {
		errorln("Run: " + err.Error())
		exit(1)
	}
}

// Run is effectively a `main` function.
// It is separated from the `main` function because of addressing an issue where` defer` is not executed when `os.Exit` is executed.
func Run(ctx context.Context) (err error) {
	flag.Parse()

	var project string
	project, err = getOptOrEnvOrDefault(optNameProjectID, *optValueProjectID, envNameGCloudProjectID, "")
	if err != nil {
		return fmt.Errorf("getOptOrEnvOrDefault: %w", err)
	}

	var dataset string
	dataset, err = getOptOrEnvOrDefault(optNameDataset, *optValueDataset, envNameBigQueryDataset, "")
	if err != nil {
		return fmt.Errorf("getOptOrEnvOrDefault: %w", err)
	}

	var filePath string
	filePath, err = getOptOrEnvOrDefault(optNameOutputFile, *optValueOutputPath, envNameOutputFile, defaultValueOutputFile)
	if err != nil {
		return fmt.Errorf("getOptOrEnvOrDefault: %w", err)
	}

	var debugString string
	debugString, err = getOptOrEnvOrDefault(optNameDebug, *optValueOutputPath, envNameDebug, defaultValueDebug)
	if err != nil {
		return fmt.Errorf("getOptOrEnvOrDefault: %w", err)
	}
	debug, _ := strconv.ParseBool(debugString)

	client, err := bigquery.NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			warnln("client.Close: " + closeErr.Error())
		}
	}()

	generatedCode, err := Generate(ctx, client, dataset, debug)
	if err != nil {
		return fmt.Errorf("Generate: %w", err)
	}

	// NOTE(ginokent): output
	if err = ioutil.WriteFile(filePath, generatedCode, 0644); err != nil {
		return fmt.Errorf("ioutil.WriteFile: %w", err)
	}

	return nil
}

func Generate(ctx context.Context, client *bigquery.Client, dataset string, debug bool) (generatedCode []byte, err error) {

	const head = `// Code generated by go run github.com/ginokent/bqschema-gen-go; DO NOT EDIT.

//go:generate go run github.com/ginokent/bqschema-gen-go

package bqschema

`

	tables, err := getAllTables(ctx, client, dataset)
	if err != nil {
		return nil, fmt.Errorf("getAllTables: %w", err)
	}

	var tail string
	var importPackages []string
	for _, table := range tables {
		var structCode string
		var pkgs []string
		structCode, pkgs, err = generateTableSchemaCode(ctx, table)
		if err != nil {
			warnln("generateTableSchemaCode: " + err.Error())
			continue
		}

		if len(pkgs) > 0 {
			importPackages = append(importPackages, pkgs...)
		}
		tail = tail + structCode
	}

	importCode := generateImportPackagesCode(importPackages)

	// NOTE(ginokent): combine
	code := head + importCode + tail

	if debug {
		fmt.Println(">>>> DEBUG >>>>>>>>>>>>>>>>")
		fmt.Println(code)
		fmt.Println("<<<< DEBUG <<<<<<<<<<<<<<<<")
	}

	gen := []byte(code)

	genFmt, err := format.Source(gen)
	if err != nil {
		return nil, fmt.Errorf("format.Source: %w", err)
	}

	if debug {
		fmt.Println(">>>> DEBUG >>>>>>>>>>>>>>>>")
		fmt.Println(string(genFmt))
		fmt.Println("<<<< DEBUG <<<<<<<<<<<<<<<<")
	}

	genImports, err := imports.Process("", genFmt, nil)
	if err != nil {
		return nil, fmt.Errorf("imports.Process: %w", err)
	}

	return genImports, nil
}

func generateImportPackagesCode(importPackages []string) (generatedCode string) {
	importPackagesUniq := make(map[string]bool)
	for _, pkg := range importPackages {
		importPackagesUniq[pkg] = true
	}

	// NOTE(ginokent): fix order
	importPackagesUniqSort := make([]string, len(importPackagesUniq))
	idx := 0
	for _, pkg := range importPackages {
		if importPackagesUniq[pkg] {
			importPackagesUniq[pkg] = false
			importPackagesUniqSort[idx] = pkg
			idx++
		}
	}

	switch {
	case len(importPackagesUniq) == 0:
		generatedCode = ""
	case len(importPackagesUniq) == 1:
		for pkg := range importPackagesUniq {
			generatedCode = "import \"" + pkg + "\"\n"
		}
		generatedCode = generatedCode + "\n"
	case len(importPackagesUniq) >= 2:
		generatedCode = "import (\n"
		for _, pkg := range importPackagesUniqSort {
			generatedCode = generatedCode + "\t\"" + pkg + "\"\n"
		}
		generatedCode = generatedCode + ")\n\n"
	}

	return generatedCode
}

func generateTableSchemaCode(ctx context.Context, table *bigquery.Table) (generatedCode string, importPackages []string, err error) {
	tableID := table.TableID
	if len(tableID) == 0 {
		return "", nil, fmt.Errorf("*bigquery.Table.TableID is empty. *bigquery.Table struct dump: %#v", table)
	}

	if strings.Contains(tableID, "-") {
		replaced := strings.ReplaceAll(tableID, "-", "_")
		warnln(fmt.Sprintf("tableID `%s` contains invalid character `-`. replacing `%s` to `%s`", tableID, tableID, replaced))
		tableID = replaced
	}

	structName := capitalizeInitial(tableID)

	var md *bigquery.TableMetadata
	md, err = table.Metadata(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("table.Metadata: %w", err)
	}

	// NOTE(ginokent): structs
	generatedCode = "// " + structName + " is BigQuery Table `" + md.FullID + "` schema struct.\n" +
		"// Description: " + md.Description + "\n" +
		"type " + structName + " struct {\n"

	schemas := []*bigquery.FieldSchema(md.Schema)

	for _, schema := range schemas {
		var goTypeStr, pkg string
		goTypeStr, pkg, err = bigqueryFieldTypeToGoType(schema.Type)
		if err != nil {
			return "", nil, fmt.Errorf("bigqueryFieldTypeToGoType: %w", err)
		}
		if pkg != "" {
			importPackages = append(importPackages, pkg)
		}
		generatedCode = generatedCode + "\t" + capitalizeInitial(schema.Name) + " " + goTypeStr + " `bigquery:\"" + schema.Name + "\"`\n"
	}
	generatedCode = generatedCode + "}\n"

	return generatedCode, importPackages, nil
}

func getAllTables(ctx context.Context, client *bigquery.Client, datasetID string) (tables []*bigquery.Table, err error) {
	tableIterator := client.Dataset(datasetID).Tables(ctx)
	for {
		var table *bigquery.Table
		table, err = tableIterator.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, fmt.Errorf("tableIterator.Next: %w", err)
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func readFile(path string) (content []byte, err error) {
	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	var bytea []byte
	bytea, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll: %w", err)
	}

	return bytea, nil
}

func getOptOrEnvOrDefault(optName, optValue, envName, defaultValue string) (value string, err error) {
	if optName == "" {
		return "", fmt.Errorf("optName is empty")
	}

	if optValue != "" {
		infoln("use option value: -" + optName + "=" + optValue)
		return optValue, nil
	}

	envValue := os.Getenv(envName)
	if envValue != "" {
		infoln("use environment variable: " + envName + "=" + envValue)
		return envValue, nil
	}

	if defaultValue != "" {
		infoln("use default option value: -" + optName + "=" + defaultValue)
		return defaultValue, nil
	}

	return "", fmt.Errorf("set option -%s, or set environment variable %s", optName, envName)
}

func capitalizeInitial(s string) (capitalized string) {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func infoln(content string) {
	log.Println("INFO: " + content)
}

func warnln(content string) {
	log.Println("WARN: " + content)
}

func errorln(content string) {
	log.Println("ERROR: " + content)
}

func exit(code int) {
	if os.Getenv("GOTEST") == "true" {
		return
	}
	os.Exit(code)
}

// NOTE(ginokent): ref. https://github.com/googleapis/google-cloud-go/blob/f37f118c87d4d0a77a554515a430ae06e5852294/bigquery/schema.go#L216
var typeOfByteSlice = reflect.TypeOf([]byte{})

// NOTE(ginokent): ref. https://github.com/googleapis/google-cloud-go/blob/f37f118c87d4d0a77a554515a430ae06e5852294/bigquery/params.go#L81-L87
var (
	typeOfDate     = reflect.TypeOf(civil.Date{})
	typeOfTime     = reflect.TypeOf(civil.Time{})
	typeOfDateTime = reflect.TypeOf(civil.DateTime{})
	typeOfGoTime   = reflect.TypeOf(time.Time{})
	typeOfRat      = reflect.TypeOf(&big.Rat{})
)

func bigqueryFieldTypeToGoType(bigqueryFieldType bigquery.FieldType) (goType string, pkg string, err error) {
	switch bigqueryFieldType {
	// NOTE(ginokent): ref. https://github.com/googleapis/google-cloud-go/blob/f37f118c87d4d0a77a554515a430ae06e5852294/bigquery/schema.go#L342-L343
	case bigquery.BytesFieldType:
		return typeOfByteSlice.String(), "", nil

	// NOTE(ginokent): ref. https://github.com/googleapis/google-cloud-go/blob/f37f118c87d4d0a77a554515a430ae06e5852294/bigquery/schema.go#L344-L358
	case bigquery.DateFieldType:
		return typeOfDate.String(), typeOfDate.PkgPath(), nil
	case bigquery.TimeFieldType:
		return typeOfTime.String(), typeOfTime.PkgPath(), nil
	case bigquery.DateTimeFieldType:
		return typeOfDateTime.String(), typeOfDateTime.PkgPath(), nil
	case bigquery.TimestampFieldType:
		return typeOfGoTime.String(), typeOfGoTime.PkgPath(), nil
	case bigquery.NumericFieldType:
		// NOTE(ginokent): The *T (pointer type) does not return the package path.
		//               ref. https://github.com/golang/go/blob/f0ff6d4a67ec9a956aa655d487543da034cf576b/src/reflect/type.go#L83
		return typeOfRat.String(), reflect.TypeOf(big.Rat{}).PkgPath(), nil

	// NOTE(ginokent): ref. https://github.com/googleapis/google-cloud-go/blob/f37f118c87d4d0a77a554515a430ae06e5852294/bigquery/schema.go#L362-L364
	case bigquery.IntegerFieldType:
		return reflect.Int64.String(), "", nil

	// NOTE(ginokent): ref. https://github.com/googleapis/google-cloud-go/blob/f37f118c87d4d0a77a554515a430ae06e5852294/bigquery/schema.go#L368-L371
	case bigquery.RecordFieldType:
		// TODO(ginokent): support bigquery.RecordFieldType
		return "", "", fmt.Errorf("bigquery.FieldType not supported. bigquery.FieldType=%s", bigqueryFieldType)

	// NOTE(ginokent): ref. https://github.com/googleapis/google-cloud-go/blob/f37f118c87d4d0a77a554515a430ae06e5852294/bigquery/schema.go#L394-L399
	case bigquery.StringFieldType, bigquery.GeographyFieldType:
		return reflect.String.String(), "", nil
	case bigquery.BooleanFieldType:
		return reflect.Bool.String(), "", nil
	case bigquery.FloatFieldType:
		return reflect.Float64.String(), "", nil

	// NOTE(ginokent): ref. https://github.com/googleapis/google-cloud-go/blob/f37f118c87d4d0a77a554515a430ae06e5852294/bigquery/schema.go#L400-L401
	default:
		return "", "", fmt.Errorf("bigquery.FieldType not supported. bigquery.FieldType=%s", bigqueryFieldType)
	}
}
