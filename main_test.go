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
	testOptName      = "test-opt-key"
	testOptValue     = "testOptValue"
	testEnvName      = "TEST_ENV_KEY"
	testEnvValue     = "testEnvValue"
	testDefaultValue = "testDefaultValue"

	// capitalizeInitial
	testNotCapitalized = "a"
	testCapitalized    = "A"

	// bigqueryFieldTypeToGoType
	testNotSupportedFieldType = "notSupportedFieldType"
)

func Test_Run_OK_1(t *testing.T) {
	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	// projectID
	backupEnvNameGCloudProjectID, exist := os.LookupEnv(envNameGCloudProjectID)
	_ = os.Setenv(envNameGCloudProjectID, testPublicDataProjectID)
	defer func() {
		if exist {
			_ = os.Setenv(envNameGCloudProjectID, backupEnvNameGCloudProjectID)
			return
		}
		_ = os.Unsetenv(envNameGCloudProjectID)
	}()

	// datasetID
	backupEnvNameBigQueryDatasetValue, exist := os.LookupEnv(envNameBigQueryDataset)
	_ = os.Setenv(envNameBigQueryDataset, testSupportedDatasetID)
	defer func() {
		if exist {
			_ = os.Setenv(envNameBigQueryDataset, backupEnvNameBigQueryDatasetValue)
			return
		}
		_ = os.Unsetenv(envNameBigQueryDataset)
	}()

	var (
		ctx = context.Background()
	)
	if err := Run(ctx); err != nil {
		t.Error(err)
	}
}

func Test_Generate_OK_1(t *testing.T) {
	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	var (
		ctx       = context.Background()
		client, _ = bigquery.NewClient(ctx, testPublicDataProjectID)
	)
	_, err := Generate(ctx, client, testSupportedDatasetID)
	if err != nil {
		t.Error(err)
	}
}

func Test_Generate_OK_2(t *testing.T) {
	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	var (
		ctx       = context.Background()
		client, _ = bigquery.NewClient(ctx, testPublicDataProjectID)
	)
	_, err := Generate(ctx, client, testNotSupportedDatasetID)
	if err != nil {
		t.Error(err)
	}
}

func Test_generateImportPackagesCode_OK_1(t *testing.T) {
	var (
		testImportsSlice = []string{}
	)
	generatedCode := generateImportPackagesCode(testImportsSlice)
	if generatedCode != testEmptyString {
		t.Error()
	}
}

func Test_generateImportPackagesCode_OK_2(t *testing.T) {
	const (
		testImportCode = "import \"time\"\n\n"
	)
	var (
		testImportsSlice = []string{"time"}
	)
	generatedCode := generateImportPackagesCode(testImportsSlice)

	if generatedCode != testImportCode {
		var (
			rr      = strings.NewReplacer("\n", "\\n", "`", "\\`")
			want    = rr.Replace(testImportCode)
			current = rr.Replace(generatedCode)
		)
		t.Error("generateImportPackagesCode: want=`" + want + "` current=`" + current + "`")
	}
}

func Test_generateImportPackagesCode_OK_3(t *testing.T) {
	const (
		testImportCode = `import (
	"math/big"
	"time"
)

`
	)
	var (
		testImportsSlice = []string{"math/big", "time"}
	)
	generatedCode := generateImportPackagesCode(testImportsSlice)

	if generatedCode != testImportCode {
		var (
			rr      = strings.NewReplacer("\n", "\\n", "`", "\\`")
			want    = rr.Replace(testImportCode)
			current = rr.Replace(generatedCode)
		)
		t.Error("generateImportPackagesCode: want=`" + want + "` current=`" + current + "`")
	}
}

func Test_generateTableSchemaCode_OK(t *testing.T) {
	var (
		ctx = context.Background()
	)

	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	var (
		ngClient, _     = bigquery.NewClient(ctx, testPublicDataProjectID)
		ngTableIterator = ngClient.Dataset(testSupportedDatasetID).Tables(ctx)
	)

	for {
		table, err := ngTableIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Error(err)
		}
		if _, _, err := generateTableSchemaCode(ctx, table); err != nil {
			t.Error(err)
		}
	}
}

func Test_generateTableSchemaCode_NG_1(t *testing.T) {
	var (
		ctx     = context.Background()
		ngTable = &bigquery.Table{
			ProjectID: testProjectNotFound,
			DatasetID: testDatasetNotFound,
			TableID:   testEmptyString,
		}
	)
	if _, _, err := generateTableSchemaCode(ctx, ngTable); err == nil {
		t.Error(err)
	}
}

func Test_generateTableSchemaCode_NG_2(t *testing.T) {
	var (
		ctx = context.Background()
	)

	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	var (
		ngClient, _ = bigquery.NewClient(ctx, testPublicDataProjectID)
		ngTable, _  = ngClient.Dataset(testNotSupportedDatasetID).Tables(ctx).Next()
	)

	ngTable.ProjectID = testProjectNotFound
	if _, _, err := generateTableSchemaCode(ctx, ngTable); err == nil {
		t.Error(err)
	}
}

func Test_generateTableSchemaCode_NG_3(t *testing.T) {
	var (
		ctx = context.Background()
	)

	if os.Getenv(envNameGoogleApplicationCredentials) == "" {
		t.Skip("WARN: " + envNameGoogleApplicationCredentials + " is not set")
	}

	var (
		ngClient, _     = bigquery.NewClient(ctx, testPublicDataProjectID)
		ngTableIterator = ngClient.Dataset(testNotSupportedDatasetID).Tables(ctx)
	)

	for {
		table, err := ngTableIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Error(err)
		}
		if _, _, err := generateTableSchemaCode(ctx, table); err != nil {
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
		ctx         = context.Background()
		okClient, _ = bigquery.NewClient(ctx, testPublicDataProjectID)
	)

	if _, err := getAllTables(ctx, okClient, testSupportedDatasetID); err != nil {
		t.Error(err)
	}
}

func Test_getAllTables_NG(t *testing.T) {
	backupValue, exist := os.LookupEnv(envNameGoogleApplicationCredentials)
	_ = os.Setenv(envNameGoogleApplicationCredentials, testGoogleApplicationCredentials)
	defer func() {
		if exist {
			_ = os.Setenv(envNameGoogleApplicationCredentials, backupValue)
			return
		}
		_ = os.Unsetenv(envNameGoogleApplicationCredentials)
	}()

	var (
		ctx         = context.Background()
		ngClient, _ = bigquery.NewClient(ctx, testProjectNotFound)
	)

	if _, err := getAllTables(ctx, ngClient, testDatasetNotFound); err == nil {
		t.Error(err)
	}
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
	v, err := getOptOrEnvOrDefault(testOptName, testOptValue, testEnvName, testDefaultValue)
	if err != nil {
		t.Error(err)
	}
	if v != testOptValue {
		t.Error(err)
	}
}
func Test_getOptOrEnvOrDefault_OK_2(t *testing.T) {
	if err := os.Setenv(testEnvName, testEnvValue); err != nil {
		t.Error(err)
	}
	v, err := getOptOrEnvOrDefault(testOptName, testEmptyString, testEnvName, testDefaultValue)
	if err != nil {
		t.Error(err)
	}
	if v != testEnvValue {
		t.Error(err)
	}
	if err := os.Unsetenv(testEnvName); err != nil {
		t.Error(err)
	}

}

func Test_getOptOrEnvOrDefault_OK_3(t *testing.T) {
	v, err := getOptOrEnvOrDefault(testOptName, testEmptyString, testEnvName, testDefaultValue)
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
	v, err := getOptOrEnvOrDefault(testOptName, testEmptyString, testEnvName, testEmptyString)
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

func Test_infoln(t *testing.T) {
	infoln("test")
}

func Test_warnln(t *testing.T) {
	warnln("test")
}

func Test_errorln(t *testing.T) {
	errorln("test")
}

func Test_exit(t *testing.T) {
	var (
		envNameGoTest  = "GOTEST"
		envValueGoTest = "true"
	)

	backupValue, exist := os.LookupEnv(envNameGoTest)
	_ = os.Setenv(envNameGoTest, envValueGoTest)
	defer func() {
		if exist {
			_ = os.Setenv(envNameGoTest, backupValue)
			return
		}
		_ = os.Unsetenv(envNameGoTest)
	}()

	exit(1)
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
