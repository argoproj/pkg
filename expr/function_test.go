package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsFloat(t *testing.T) {
	assert.Equal(t, int64(1), AsInt(int32(1)))
	assert.Equal(t, int64(1), AsInt(int64(1)))
	assert.Equal(t, int64(1), AsInt("1"))
	assert.Panics(t, func() { AsInt("1.56") })
}

func TestAsFloat2(t *testing.T) {
	assert.Equal(t, 1.24, AsFloat(1.24))
	assert.Equal(t, 1.65, AsFloat("1.65"))
}

func TestAsStr(t *testing.T) {
	assert.Equal(t, "1.24", AsStr(1.24))
	assert.Equal(t, "1", AsStr(1))
}

func TestJsonPath(t *testing.T) {
	simpleJson := "{\"employee\":{\"name\":\"sonoo\",\"salary\":56000,\"married\":true}}"
	arrayJson := "{\"employees\":[{\"name\":\"Shyam\",\"email\":\"shyamjaiswal@gmail.com\",\"age\":43},{\"name\":\"Bob\",\"email\":\"bob32@gmail.com\",\"age\":42}," +
		"{\"name\":\"Jai\",\"email\":\"jai87@gmail.com\",\"age\":44}]}"

	assert.Equal(t, "sonoo", JsonPath(simpleJson, "$.employee.name"))
	assert.Equal(t, "Bob", JsonPath(arrayJson, "$.employees[1].name"))
	assert.Equal(t, []interface{}{"Bob"}, JsonPath(arrayJson, `$.employees[?(@.name == "Bob")].name`))
	assert.Equal(t, []interface{}{}, JsonPath(arrayJson, `$..[?(@.name.length == 4)].name`))
	assert.Equal(t, []interface{}{"Shyam", "Bob", "Jai"}, JsonPath(arrayJson, `$..[0:4].name`))
	assert.Equal(t, []interface{}{"Shyam"}, JsonPath(arrayJson, `$.employees[0:1].name`))
	assert.Equal(t, "Jai", JsonPath(arrayJson, `$.employees[-1].name`))
	assert.Equal(t, []interface{}{"Bob", "Jai"}, JsonPath(arrayJson, `$.employees[-2:].name`))
	assert.Equal(t, []interface{}{"Bob", "Jai"}, JsonPath(arrayJson, `$.employees[1:].name`))
	assert.Equal(t, []interface{}{"Shyam", "Bob", "Jai"}, JsonPath(arrayJson, `$.employees[:].name`))
	assert.Equal(t, []interface{}{"Jai"}, JsonPath(arrayJson, `$..[2].name`))
	assert.Equal(t, []interface{}{"Shyam", "Bob", "Jai"}, JsonPath(arrayJson, `$..[*].name`))
	assert.Equal(t, []interface{}{"Bob"}, JsonPath(arrayJson, `$.employees[?(@.name=="Bob")].name`))
	assert.Equal(t, []interface{}{"Shyam", "Jai"}, JsonPath(arrayJson, `$.employees[0,2].name`))
	assert.Equal(t, []interface{}{"Shyam"}, JsonPath(arrayJson, `$.employees[?(@.age>42 && @.age<44)].name`))
	assert.Equal(t, []interface{}{"Shyam", "Jai"}, JsonPath(arrayJson, `$.employees[?(@.age!=42)].name`))
}
