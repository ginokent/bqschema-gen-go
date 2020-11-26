package main

import (
	"testing"

	"cloud.google.com/go/bigquery"
)

func Test_bigqueryFieldTypeToGoType(t *testing.T) {
	supportedBigqueryFieldTypes := map[bigquery.FieldType]inter{}{
		bigquery.StringFieldType,
		string(bigquery.BytesFieldType),
		string(bigquery.IntegerFieldType),
		string(bigquery.FloatFieldType),
		string(bigquery.BooleanFieldType),
		string(bigquery.TimestampFieldType),
		//string(bigquery.RecordFieldType),
		string(bigquery.DateFieldType),
		string(bigquery.TimeFieldType),
		string(bigquery.DateTimeFieldType),
		string(bigquery.NumericFieldType),
		string(bigquery.GeographyFieldType),
	}

	unsupportedBigqueryFieldTypes := []string{
		string(bigquery.RecordFieldType),
	}

	for _, fieldType := range supportedBigqueryFieldTypes {
		bigqueryFieldTypeToGoType(fieldType)
	}

	for _, fieldType := range unsupportedBigqueryFieldTypes {

	}
}
