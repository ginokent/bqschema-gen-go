package main

import (
	"testing"

	"cloud.google.com/go/bigquery"
)

func Test_bigqueryFieldTypeToGoType(t *testing.T) {
	supportedBigqueryFieldTypes := map[bigquery.FieldType]string{
		bigquery.StringFieldType:    typeOfString,
		bigquery.BytesFieldType:     typeOfByteSlice,
		bigquery.IntegerFieldType:   typeOfInt64,
		bigquery.FloatFieldType:     typeOfFloat64,
		bigquery.BooleanFieldType:   typeOfBool,
		bigquery.TimestampFieldType: typeOfGoTime,
		//bigquery.RecordFieldType: ,
		bigquery.DateFieldType:      typeOfDate,
		bigquery.TimeFieldType:      typeOfTime,
		bigquery.DateTimeFieldType:  typeOfDateTime,
		bigquery.NumericFieldType:   typeOfRat,
		bigquery.GeographyFieldType: typeOfString,
	}

	unsupportedBigqueryFieldTypes := map[bigquery.FieldType]string{
		bigquery.RecordFieldType: "",
	}

	for fieldType, typeString := range supportedBigqueryFieldTypes {
		goType, _, err := bigqueryFieldTypeToGoType(fieldType)
		if err != nil {
			t.Fail()
		}
		if goType != typeString {
			t.Fail()
		}
	}

	for fieldType, typeString := range unsupportedBigqueryFieldTypes {
		goType, _, err := bigqueryFieldTypeToGoType(fieldType)
		if err == nil {
			t.Fail()
		}
		if goType != typeString {
			t.Fail()
		}

	}
}
