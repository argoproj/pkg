package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDisallowUnknownFields tests ability to disallow unknown fields
func TestDisallowUnknownFields(t *testing.T) {
	type mystruct struct {
		MyField string `json:"myField"`
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

	obj = mystruct{}
	err = UnmarshalStrict(jsonWithUnknownField, &obj)
	assert.Error(t, err)
	assert.Equal(t, "foo", obj.MyField)
}

func TestIsJSON(t *testing.T) {
	assert.True(t, IsJSON([]byte(`"foo"`)))
	assert.True(t, IsJSON([]byte(`{"a": "b"}`)))
	assert.True(t, IsJSON([]byte(`[{"a": "b"}]`)))
	assert.False(t, IsJSON([]byte(`foo`)))
	assert.False(t, IsJSON([]byte(`foo: bar`)))
}
