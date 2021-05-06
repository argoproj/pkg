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
	arrayJson := "{\"employees\":[{\"name\":\"Shyam\",\"email\":\"shyamjaiswal@gmail.com\"},{\"name\":\"Bob\",\"email\":\"bob32@gmail.com\"}," +
		"{\"name\":\"Jai\",\"email\":\"jai87@gmail.com\"}]}"
	assert.Equal(t, "sonoo", JsonPath(simpleJson, "$.employee.name"))
	assert.Equal(t, "Bob", JsonPath(arrayJson, "$.employees[1].name"))
}
