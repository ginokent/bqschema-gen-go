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
	testEmptyString                = ""
	GOOGLE_APPLICATION_CREDENTIALS = "GOOGLE_APPLICATION_CREDENTIALS"

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

func Test_Run(t *testing.T) {
	t.Run("正常系_testPublicDataProjectID_"+testPublicDataProjectID+"_testSupportedDatasetID_"+testSupportedDatasetID, func(t *testing.T) {
		if os.Getenv(GOOGLE_APPLICATION_CREDENTIALS) == "" {
			t.Skip("WARN: " + GOOGLE_APPLICATION_CREDENTIALS + " is not set")
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
	})
}

func Test_Generate(t *testing.T) {
	t.Run("正常系_testSupportedDatasetID_"+testSupportedDatasetID, func(t *testing.T) {
		if os.Getenv(GOOGLE_APPLICATION_CREDENTIALS) == "" {
			t.Skip("WARN: " + GOOGLE_APPLICATION_CREDENTIALS + " is not set")
		}

		var (
			ctx       = context.Background()
			client, _ = bigquery.NewClient(ctx, testPublicDataProjectID)
		)

		_, err := Generate(ctx, client, testSupportedDatasetID)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("正常系_testNotSupportedDatasetID_"+testNotSupportedDatasetID, func(t *testing.T) {
		if os.Getenv(GOOGLE_APPLICATION_CREDENTIALS) == "" {
			t.Skip("WARN: " + GOOGLE_APPLICATION_CREDENTIALS + " is not set")
		}

		var (
			ctx       = context.Background()
			client, _ = bigquery.NewClient(ctx, testPublicDataProjectID)
		)

		_, err := Generate(ctx, client, testNotSupportedDatasetID)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_generateImportPackagesCode(t *testing.T) {
	t.Run("正常系_import_nothing", func(t *testing.T) {
		const (
			// 正しい出力
			testImportCode = ""
		)
		var (
			testImportsSlice = []string{}
		)

		generatedCode := generateImportPackagesCode(testImportsSlice)
		if generatedCode != testImportCode {
			t.Error()
		}
	})

	t.Run("正常系_import_time", func(t *testing.T) {
		const (
			// 正しい出力
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
	})

	t.Run("正常系_import_math/big_time", func(t *testing.T) {
		const (
			// 正しい出力
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
	})
}

func Test_generateTableSchemaCode(t *testing.T) {
	t.Run("正常系_testPublicDataProjectID_testPublicDataProjectID", func(t *testing.T) {
		var (
			ctx = context.Background()
		)

		if os.Getenv(GOOGLE_APPLICATION_CREDENTIALS) == "" {
			t.Skip("WARN: " + GOOGLE_APPLICATION_CREDENTIALS + " is not set")
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
	})

	t.Run("異常系_testProjectNotFound_testDatasetNotFound", func(t *testing.T) {
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
	})

	t.Run("異常系_testProjectNotFound_testNotSupportedDatasetID", func(t *testing.T) {
		var (
			ctx = context.Background()
		)

		if os.Getenv(GOOGLE_APPLICATION_CREDENTIALS) == "" {
			t.Skip("WARN: " + GOOGLE_APPLICATION_CREDENTIALS + " is not set")
		}

		var (
			ngClient, _ = bigquery.NewClient(ctx, testPublicDataProjectID)
			ngTable, _  = ngClient.Dataset(testNotSupportedDatasetID).Tables(ctx).Next()
		)

		ngTable.ProjectID = testProjectNotFound
		if _, _, err := generateTableSchemaCode(ctx, ngTable); err == nil {
			t.Error(err)
		}
	})

	t.Run("異常系_testPublicDataProjectID_testNotSupportedDatasetID_testSubStrFieldTypeNotSupported", func(t *testing.T) {
		var (
			ctx = context.Background()
		)

		if os.Getenv(GOOGLE_APPLICATION_CREDENTIALS) == "" {
			t.Skip("WARN: " + GOOGLE_APPLICATION_CREDENTIALS + " is not set")
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
				// NOTE(ginokent): "bigquery.FieldType not supported." 以外のエラーが出たら Fail
				if !strings.Contains(err.Error(), testSubStrFieldTypeNotSupported) {
					t.Error(err)
				}
				// NOTE(ginokent): ここまで来たら、確認したいことは確認済み。
				// ref. https://github.com/ginokent/bqschema-gen-go/blob/260524ce0ae2dd5bdcbdd57446cdd8c140326ca4/main.go#L212
				return
			}
		}
	})
}

func Test_getAllTables(t *testing.T) {
	t.Run("正常系_testPublicDataProjectID_testSupportedDatasetID", func(t *testing.T) {

		if os.Getenv(GOOGLE_APPLICATION_CREDENTIALS) == "" {
			t.Skip("WARN: " + GOOGLE_APPLICATION_CREDENTIALS + " is not set")
		}

		var (
			ctx         = context.Background()
			okClient, _ = bigquery.NewClient(ctx, testPublicDataProjectID)
		)

		if _, err := getAllTables(ctx, okClient, testSupportedDatasetID); err != nil {
			t.Error(err)
		}
	})

	t.Run("異常系_testProjectNotFound_testDatasetNotFound", func(t *testing.T) {

		backupValue, exist := os.LookupEnv(GOOGLE_APPLICATION_CREDENTIALS)
		_ = os.Setenv(GOOGLE_APPLICATION_CREDENTIALS, testGoogleApplicationCredentials)
		defer func() {
			if exist {
				_ = os.Setenv(GOOGLE_APPLICATION_CREDENTIALS, backupValue)
				return
			}
			_ = os.Unsetenv(GOOGLE_APPLICATION_CREDENTIALS)
		}()

		var (
			ctx         = context.Background()
			ngClient, _ = bigquery.NewClient(ctx, testProjectNotFound)
		)

		if _, err := getAllTables(ctx, ngClient, testDatasetNotFound); err == nil {
			t.Error(err)
		}
	})
}

func Test_readFile(t *testing.T) {
	t.Run("正常系_testProbablyExistsPath", func(t *testing.T) {
		if _, err := readFile(testProbablyExistsPath); err != nil {
			t.Error(err)
		}
	})

	t.Run("異常系_testErrNoSuchFileOrDirectoryPath", func(t *testing.T) {
		if _, err := readFile(testErrNoSuchFileOrDirectoryPath); err == nil {
			t.Error(err)
		}
	})

	t.Run("異常系_testErrIsADirectoryPath", func(t *testing.T) {
		if _, err := readFile(testErrIsADirectoryPath); err == nil {
			t.Error(err)
		}
	})
}

func Test_getOptOrEnvOrDefault(t *testing.T) {
	t.Run("正常系_testOptValue", func(t *testing.T) {
		v, err := getOptOrEnvOrDefault(testOptName, testOptValue, testEnvName, testDefaultValue)
		if err != nil {
			t.Error(err)
		}
		if v != testOptValue {
			t.Error(err)
		}
	})

	t.Run("正常系_testEnvValue", func(t *testing.T) {
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
	})

	t.Run("正常系_testDefaultValue", func(t *testing.T) {
		v, err := getOptOrEnvOrDefault(testOptName, testEmptyString, testEnvName, testDefaultValue)
		if err != nil {
			t.Error(err)
		}
		if v != testDefaultValue {
			t.Error(err)
		}
	})

	t.Run("異常系_testEmptyString_all", func(t *testing.T) {
		v, err := getOptOrEnvOrDefault(testEmptyString, testEmptyString, testEmptyString, testEmptyString)
		if err == nil {
			t.Error(err)
		}
		if v != testEmptyString {
			t.Error(err)
		}
	})

	t.Run("異常系_testEmptyString", func(t *testing.T) {
		v, err := getOptOrEnvOrDefault(testOptName, testEmptyString, testEnvName, testEmptyString)
		if err == nil {
			t.Error(err)
		}
		if v != testEmptyString {
			t.Error(err)
		}
	})
}

func Test_capitalizeInitial(t *testing.T) {
	t.Run("正常系_testEmptyString", func(t *testing.T) {
		if capitalizeInitial(testEmptyString) != testEmptyString {
			t.Error()
		}
	})

	t.Run("正常系_testCapitalized", func(t *testing.T) {
		if capitalizeInitial(testNotCapitalized) != testCapitalized {
			t.Error()
		}
	})
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
			// TODO(ginokent): support bigquery.RecordFieldType
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

	t.Run("正常系_supportedBigqueryFieldTypes", func(t *testing.T) {
		for bigqueryFieldType, typeOf := range supportedBigqueryFieldTypes {
			goType, _, err := bigqueryFieldTypeToGoType(bigqueryFieldType)
			if err != nil {
				t.Error(err)
			}
			if goType != typeOf {
				t.Error()
			}
		}
	})

	t.Run("異常系_unsupportedBigqueryFieldTypes", func(t *testing.T) {
		for bigqueryFieldType, typeOf := range unsupportedBigqueryFieldTypes {
			goType, _, err := bigqueryFieldTypeToGoType(bigqueryFieldType)
			if err == nil {
				t.Error(err)
			}
			if goType != typeOf {
				t.Error()
			}
		}
	})
}
