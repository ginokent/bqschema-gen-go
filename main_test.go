package main

import (
	"reflect"
	"testing"

	"cloud.google.com/go/bigquery"
)

func Test_bigqueryFieldTypeToGoType(t *testing.T) {
	supportedBigqueryFieldTypes := map[bigquery.FieldType]string{
		bigquery.StringFieldType:    reflect.String.String(),
		bigquery.BytesFieldType:     typeOfByteSlice.String(),
		bigquery.IntegerFieldType:   reflect.Int64.String(),
		bigquery.FloatFieldType:     reflect.Float64.String(),
		bigquery.BooleanFieldType:   reflect.Bool.String(),
		bigquery.TimestampFieldType: typeOfGoTime.String(),
		//bigquery.RecordFieldType: "",
		bigquery.DateFieldType:      typeOfDate.String(),
		bigquery.TimeFieldType:      typeOfTime.String(),
		bigquery.DateTimeFieldType:  typeOfDateTime.String(),
		bigquery.NumericFieldType:   typeOfRat.String(),
		bigquery.GeographyFieldType: reflect.String.String(),
	}

	unsupportedBigqueryFieldTypes := map[bigquery.FieldType]string{
		bigquery.RecordFieldType: "",
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
