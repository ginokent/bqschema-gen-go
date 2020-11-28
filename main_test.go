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
	// all
	testEmptyString = ""

	// generateTableSchemaCode, getAllTables
	testPublicDataProjectID         = "bigquery-public-data"
	testSupportedDatasetID          = "hacker_news"
	testNotSupportedDatasetID       = "samples"
	testProjectNotFound             = "projectnotfound"
	testDatasetNotFound             = "datasetnotfound"
	testSubStrFieldTypeNotSupported = "bigquery.FieldType not supported."

	// getAllTables
	testGoogleApplicationCredentials = "test/serviceaccountnotfound@projectnotfound.iam.gserviceaccount.com.json"

	// readFile
	testErrNoSuchFileOrDirectoryPath = "/no/such/file/or/directory"
	testErrIsADirectoryPath          = "."
	testProbablyExistsPath           = "go.mod"

	// getOptOrEnvOrDefault
	testOptKey       = "testOptKey"
	testOptValue     = "testOptValue"
	testEnvKey       = "TEST_ENV_KEY"
	testEnvValue     = "testEnvValue"
	testDefaultValue = "testDefaultValue"

	// capitalizeInitial
	testNotCapitalized = "a"
	testCapitalized    = "A"

	// bigqueryFieldTypeToGoType
	testNotSupportedFieldType = "notSupportedFieldType"
)

func Test_generateImportPackagesCode_OK_1(t *testing.T) {
	var (
		testEmptySlice = []string{}
	)
	generatedCode := generateImportPackagesCode(testEmptySlice)
	if generatedCode != testEmptyString {
		t.Error()
	}
}

func Test_generateTableSchemaCode_OK(t *testing.T) {
	var (
		okCtx = context.Background()
	)

	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	var (
		ngClient, _     = bigquery.NewClient(okCtx, testPublicDataProjectID)
		ngTableIterator = ngClient.Dataset(testSupportedDatasetID).Tables(okCtx)
	)

	for {
		table, err := ngTableIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Error(err)
		}
		if _, _, err := generateTableSchemaCode(okCtx, table); err != nil {
			t.Error(err)
		}
	}
}

func Test_generateTableSchemaCode_NG_1(t *testing.T) {
	var (
		ngCtx   = context.Background()
		ngTable = &bigquery.Table{
			ProjectID: testProjectNotFound,
			DatasetID: testDatasetNotFound,
			TableID:   testEmptyString,
		}
	)
	if _, _, err := generateTableSchemaCode(ngCtx, ngTable); err == nil {
		t.Error(err)
	}
}

func Test_generateTableSchemaCode_NG_2(t *testing.T) {
	var (
		ngCtx = context.Background()
	)

	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	var (
		ngClient, _ = bigquery.NewClient(ngCtx, testPublicDataProjectID)
		ngTable, _  = ngClient.Dataset(testNotSupportedDatasetID).Tables(ngCtx).Next()
	)

	ngTable.ProjectID = testProjectNotFound
	if _, _, err := generateTableSchemaCode(ngCtx, ngTable); err == nil {
		t.Error(err)
	}
}

func Test_generateTableSchemaCode_NG_3(t *testing.T) {
	var (
		ngCtx = context.Background()
	)

	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	var (
		ngClient, _     = bigquery.NewClient(ngCtx, testPublicDataProjectID)
		ngTableIterator = ngClient.Dataset(testNotSupportedDatasetID).Tables(ngCtx)
	)

	for {
		table, err := ngTableIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Error(err)
		}
		if _, _, err := generateTableSchemaCode(ngCtx, table); err != nil {
			// NOTE(djeeno): "bigquery.FieldType not supported." 以外のエラーが出たら Fail
			if !strings.Contains(err.Error(), testSubStrFieldTypeNotSupported) {
				t.Error(err)
			}
			// NOTE(djeeno): ここまで来たら、確認したいことは確認済み。
			// ref. https://github.com/djeeno/bqtableschema/blob/260524ce0ae2dd5bdcbdd57446cdd8c140326ca4/main.go#L212
			return
		}
	}
}

func Test_getAllTables_OK(t *testing.T) {
	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	var (
		okCtx       = context.Background()
		okClient, _ = bigquery.NewClient(okCtx, testPublicDataProjectID)
	)

	if _, err := getAllTables(okCtx, okClient, testSupportedDatasetID); err != nil {
		t.Error(err)
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
		ngClient, _ = bigquery.NewClient(ngCtx, testProjectNotFound)
	)

	if _, err := getAllTables(ngCtx, ngClient, testDatasetNotFound); err == nil {
		t.Error(err)
	}

	if exist {
		_ = os.Setenv(envNameGoogleApplicationCredentials, backupValue)
		return
	}

	_ = os.Unsetenv(envNameGoogleApplicationCredentials)
}

func Test_readFile_OK(t *testing.T) {
	if _, err := readFile(testProbablyExistsPath); err != nil {
		t.Error(err)
	}
}

func Test_readFile_NG_1(t *testing.T) {
	if _, err := readFile(testErrNoSuchFileOrDirectoryPath); err == nil {
		t.Error(err)
	}
}

func Test_readFile_NG_2(t *testing.T) {
	if _, err := readFile(testErrIsADirectoryPath); err == nil {
		t.Error(err)
	}
}

func Test_getOptOrEnvOrDefault_OK_1(t *testing.T) {
	v, err := getOptOrEnvOrDefault(testOptKey, testOptValue, testEnvKey, testDefaultValue)
	if err != nil {
		t.Error(err)
	}
	if v != testOptValue {
		t.Error(err)
	}
}
func Test_getOptOrEnvOrDefault_OK_2(t *testing.T) {
	if err := os.Setenv(testEnvKey, testEnvValue); err != nil {
		t.Error(err)
	}
	v, err := getOptOrEnvOrDefault(testOptKey, testEmptyString, testEnvKey, testDefaultValue)
	if err != nil {
		t.Error(err)
	}
	if v != testEnvValue {
		t.Error(err)
	}
	if err := os.Unsetenv(testEnvKey); err != nil {
		t.Error(err)
	}

}

func Test_getOptOrEnvOrDefault_OK_3(t *testing.T) {
	v, err := getOptOrEnvOrDefault(testOptKey, testEmptyString, testEnvKey, testDefaultValue)
	if err != nil {
		t.Error(err)
	}
	if v != testDefaultValue {
		t.Error(err)
	}
}

func Test_getOptOrEnvOrDefault_NG_1(t *testing.T) {
	v, err := getOptOrEnvOrDefault(testEmptyString, testEmptyString, testEmptyString, testEmptyString)
	if err == nil {
		t.Error(err)
	}
	if v != testEmptyString {
		t.Error(err)
	}
}

func Test_getOptOrEnvOrDefault_NG_2(t *testing.T) {
	v, err := getOptOrEnvOrDefault(testOptKey, testEmptyString, testEnvKey, testEmptyString)
	if err == nil {
		t.Error(err)
	}
	if v != testEmptyString {
		t.Error(err)
	}
}

func Test_capitalizeInitial(t *testing.T) {
	if capitalizeInitial(testEmptyString) != testEmptyString {
		t.Error()
	}

	if capitalizeInitial(testNotCapitalized) != testCapitalized {
		t.Error()
	}
}

func Test_bigqueryFieldTypeToGoType(t *testing.T) {
	var (
		supportedBigqueryFieldTypes = map[bigquery.FieldType]string{
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

		unsupportedBigqueryFieldTypes = map[bigquery.FieldType]string{
			bigquery.RecordFieldType:                      testEmptyString,
			bigquery.FieldType(testNotSupportedFieldType): testEmptyString,
		}
	)

	for bigqueryFieldType, typeOf := range supportedBigqueryFieldTypes {
		goType, _, err := bigqueryFieldTypeToGoType(bigqueryFieldType)
		if err != nil {
			t.Error(err)
		}
		if goType != typeOf {
			t.Error()
		}
	}

	for bigqueryFieldType, typeOf := range unsupportedBigqueryFieldTypes {
		goType, _, err := bigqueryFieldTypeToGoType(bigqueryFieldType)
		if err == nil {
			t.Error(err)
		}
		if goType != typeOf {
			t.Error()
		}
	}
}
