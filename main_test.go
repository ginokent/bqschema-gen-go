package main

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
	testGoogleApplicationCredentials = "test/serviceaccountnotfound@projectnotfound.iam.gserviceaccount.com.json"
)

func Test_generateTableSchemaCode_OK(t *testing.T) {
	var (
		okCtx       = context.Background()
		okProjectID = "bigquery-public-data"
		okDatasetID = "hacker_news"
	)

	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Log("WARN: " + envNameGoogleApplicationCredentials + " is not set")
		return
	}

	var (
		ngClient, _     = bigquery.NewClient(okCtx, okProjectID)
		ngTableIterator = ngClient.Dataset(okDatasetID).Tables(okCtx)
	)

	for {
		table, err := ngTableIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if _, _, err := generateTableSchemaCode(okCtx, table); err != nil {
			t.Log(err)
			t.Fail()
		}
	}
}

func Test_generateTableSchemaCode_NG_1(t *testing.T) {
	var (
		ngCtx   = context.Background()
		ngTable = &bigquery.Table{
			ProjectID: "",
			DatasetID: "",
			TableID:   "",
		}
	)
	if _, _, err := generateTableSchemaCode(ngCtx, ngTable); err == nil {
		t.Log(err)
		t.Fail()
	}
}

func Test_generateTableSchemaCode_NG_2(t *testing.T) {
	var (
		ngCtx       = context.Background()
		ngProjectID = "bigquery-public-data"
		ngDatasetID = "samples"
	)

	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Log("WARN: " + envNameGoogleApplicationCredentials + " is not set")
		return
	}

	var (
		ngClient, _ = bigquery.NewClient(ngCtx, ngProjectID)
		ngTable, _  = ngClient.Dataset(ngDatasetID).Tables(ngCtx).Next()
	)

	ngTable.ProjectID = "projectnotfound"
	if _, _, err := generateTableSchemaCode(ngCtx, ngTable); err == nil {
		t.Log(err)
		t.Fail()
	}
}

func Test_generateTableSchemaCode_NG_3(t *testing.T) {
	var (
		ngCtx       = context.Background()
		ngProjectID = "bigquery-public-data"
		ngDatasetID = "samples"
	)

	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Log("WARN: " + envNameGoogleApplicationCredentials + " is not set")
		return
	}

	var (
		ngClient, _     = bigquery.NewClient(ngCtx, ngProjectID)
		ngTableIterator = ngClient.Dataset(ngDatasetID).Tables(ngCtx)
	)

	for {
		table, err := ngTableIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if _, _, err := generateTableSchemaCode(ngCtx, table); err != nil {
			// NOTE(djeeno): "bigquery.FieldType not supported." 以外のエラーが出たら Fail
			if !strings.Contains(err.Error(), "bigquery.FieldType not supported.") {
				t.Log(err)
				t.Fail()
			}
			// NOTE(djeeno): ここまで来たら、確認したいことは確認済み。
			// ref. https://github.com/djeeno/bqtableschema/blob/260524ce0ae2dd5bdcbdd57446cdd8c140326ca4/main.go#L212
			return
		}
	}
}

func Test_getAllTables_OK(t *testing.T) {
	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Log("WARN: " + envNameGoogleApplicationCredentials + " is not set")
		return
	}

	var (
		okCtx       = context.Background()
		okProjectID = "bigquery-public-data"
		okClient, _ = bigquery.NewClient(okCtx, okProjectID)
		okDatasetID = "samples"
	)

	if _, err := getAllTables(okCtx, okClient, okDatasetID); err != nil {
		t.Log(err)
		t.Fail()
	}
}

func Test_getAllTables_NG(t *testing.T) {
	var backupValue string

	v, exist := os.LookupEnv(envNameGoogleApplicationCredentials)
	if exist {
		backupValue = v
	}

	_ = os.Setenv(envNameGoogleApplicationCredentials, testGoogleApplicationCredentials)

	var (
		ngCtx       = context.Background()
		ngProjectID = "projectnotfound"
		ngClient, _ = bigquery.NewClient(ngCtx, ngProjectID)
		ngDatasetID = "datasetnotfound"
	)

	if _, err := getAllTables(ngCtx, ngClient, ngDatasetID); err == nil {
		t.Log(err)
		t.Fail()
	}

	if exist {
		_ = os.Setenv(envNameGoogleApplicationCredentials, backupValue)
		return
	}

	_ = os.Unsetenv(envNameGoogleApplicationCredentials)
	return
}

func Test_newGoogleApplicationCredentials(t *testing.T) {
	var (
		noSuchFileOrDirectoryPath = "/no/such/file/or/directory"
		cannotJSONMarshalPath     = "go.mod"
	)

	if _, err := newGoogleApplicationCredentials(noSuchFileOrDirectoryPath); err == nil {
		t.Log(err)
		t.Fail()
	}

	if _, err := newGoogleApplicationCredentials(cannotJSONMarshalPath); err == nil {
		t.Log(err)
		t.Fail()
	}

	if _, err := newGoogleApplicationCredentials(testGoogleApplicationCredentials); err != nil {
		t.Log(err)
		t.Fail()
	}
}

func Test_readFile(t *testing.T) {
	var (
		errNoSuchFileOrDirectoryPath = "/no/such/file/or/directory"
		errIsADirectory              = "."
		probablyExistsPath           = "go.mod"
	)
	if _, err := readFile(errNoSuchFileOrDirectoryPath); err == nil {
		t.Log(err)
		t.Fail()
	}

	if _, err := readFile(errIsADirectory); err == nil {
		t.Log(err)
		t.Fail()
	}

	if _, err := readFile(probablyExistsPath); err != nil {
		t.Log(err)
		t.Fail()
	}
}

func Test_getOptOrEnvOrDefault(t *testing.T) {
	var (
		empty            = ""
		testOptKey       = "testOptKey"
		testOptValue     = "testOptValue"
		testEnvKey       = "TEST_ENV_KEY"
		testEnvValue     = "testEnvValue"
		testDefaultValue = "testDefaultValue"
	)

	{
		v, err := getOptOrEnvOrDefault(empty, empty, empty, empty)
		if err == nil {
			t.Log(err)
			t.Fail()
		}
		if v != empty {
			t.Log(err)
			t.Fail()
		}
	}

	{
		v, err := getOptOrEnvOrDefault(testOptKey, testOptValue, testEnvKey, testDefaultValue)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if v != testOptValue {
			t.Log(err)
			t.Fail()
		}
	}

	{
		_ = os.Setenv(testEnvKey, testEnvValue)
		v, err := getOptOrEnvOrDefault(testOptKey, empty, testEnvKey, testDefaultValue)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if v != testEnvValue {
			t.Log(err)
			t.Fail()
		}
		_ = os.Unsetenv(testEnvKey)
	}

	{
		v, err := getOptOrEnvOrDefault(testOptKey, empty, testEnvKey, testDefaultValue)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if v != testDefaultValue {
			t.Log(err)
			t.Fail()
		}
	}

	{
		v, err := getOptOrEnvOrDefault(testOptKey, empty, testEnvKey, empty)
		if err == nil {
			t.Log(err)
			t.Fail()
		}
		if v != empty {
			t.Log(err)
			t.Fail()
		}
	}
}

func Test_capitalizeInitial(t *testing.T) {
	var (
		empty          = ""
		notCapitalized = "a"
		capitalized    = "A"
	)

	if capitalizeInitial(empty) != empty {
		t.Log()
		t.Fail()
	}

	if capitalizeInitial(notCapitalized) != capitalized {
		t.Log()
		t.Fail()
	}
}

func Test_bigqueryFieldTypeToGoType(t *testing.T) {
	supportedBigqueryFieldTypes := map[bigquery.FieldType]string{
		bigquery.StringFieldType:    reflect.String.String(),
		bigquery.BytesFieldType:     typeOfByteSlice.String(),
		bigquery.IntegerFieldType:   reflect.Int64.String(),
		bigquery.FloatFieldType:     reflect.Float64.String(),
		bigquery.BooleanFieldType:   reflect.Bool.String(),
		bigquery.TimestampFieldType: typeOfGoTime.String(),
		// TODO(djeeno): support bigquery.RecordFieldType
		//bigquery.RecordFieldType: "",
		bigquery.DateFieldType:      typeOfDate.String(),
		bigquery.TimeFieldType:      typeOfTime.String(),
		bigquery.DateTimeFieldType:  typeOfDateTime.String(),
		bigquery.NumericFieldType:   typeOfRat.String(),
		bigquery.GeographyFieldType: reflect.String.String(),
	}

	unsupportedBigqueryFieldTypes := map[bigquery.FieldType]string{
		bigquery.RecordFieldType:               "",
		bigquery.FieldType("unknownFieldType"): "",
	}

	for bigqueryFieldType, typeOf := range supportedBigqueryFieldTypes {
		goType, _, err := bigqueryFieldTypeToGoType(bigqueryFieldType)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if goType != typeOf {
			t.Log()
			t.Fail()
		}
	}

	for bigqueryFieldType, typeOf := range unsupportedBigqueryFieldTypes {
		goType, _, err := bigqueryFieldTypeToGoType(bigqueryFieldType)
		if err == nil {
			t.Log(err)
			t.Fail()
		}
		if goType != typeOf {
			t.Log()
			t.Fail()
		}

	}
}
