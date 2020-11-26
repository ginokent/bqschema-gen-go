package main

import (
	"os"
	"reflect"
	"testing"

	"cloud.google.com/go/bigquery"
)

func Test_readFile(t *testing.T) {
	var (
		noSuchFileOrDirectoryPath = "/no/such/file/or/directory"
		probablyExistsPath        = "go.mod"
	)
	if _, err := readFile(noSuchFileOrDirectoryPath); err == nil {
		t.Fail()
	}

	if _, err := readFile(probablyExistsPath); err != nil {
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
			t.Fail()
		}
		if v != empty {
			t.Fail()
		}
	}

	{
		v, err := getOptOrEnvOrDefault(testOptKey, testOptValue, testEnvKey, testDefaultValue)
		if err != nil {
			t.Fail()
		}
		if v != testOptValue {
			t.Fail()
		}
	}

	{
		_ = os.Setenv(testEnvKey, testEnvValue)
		v, err := getOptOrEnvOrDefault(testOptKey, empty, testEnvKey, testDefaultValue)
		if err != nil {
			t.Fail()
		}
		if v != testEnvValue {
			t.Fail()
		}
		_ = os.Unsetenv(testEnvKey)
	}

	{
		v, err := getOptOrEnvOrDefault(testOptKey, empty, testEnvKey, testDefaultValue)
		if err != nil {
			t.Fail()
		}
		if v != testDefaultValue {
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
		t.Fail()
	}

	if capitalizeInitial(notCapitalized) != capitalized {
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
			t.Fail()
		}
		if goType != typeOf {
			t.Fail()
		}
	}

	for bigqueryFieldType, typeOf := range unsupportedBigqueryFieldTypes {
		goType, _, err := bigqueryFieldTypeToGoType(bigqueryFieldType)
		if err == nil {
			t.Fail()
		}
		if goType != typeOf {
			t.Fail()
		}

	}
}
