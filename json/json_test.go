package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDisallowUnknownFields tests ability to disallow unknown fields
func TestDisallowUnknownFields(t *testing.T) {
	type mystruct struct {
		MyField string `json:myField`
	}
	jsonWithUnknownField := []byte(`
	{
		"myField": "foo",
		"unknown": "bar"
	}
	`)
	var obj mystruct

	err := Unmarshal(jsonWithUnknownField, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "foo", obj.MyField)

	obj = mystruct{}
	err = Unmarshal(jsonWithUnknownField, &obj, DisallowUnknownFields)
	assert.Error(t, err)
	assert.Equal(t, "foo", obj.MyField)
}
