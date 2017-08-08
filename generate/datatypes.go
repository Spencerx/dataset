package generate

import (
	"math/rand"
	"net/url"
	"time"

	"github.com/qri-io/dataset/datatypes"
)

// RandomValue generates a randomized value for a given datatype
func RandomValue(t datatypes.Type) interface{} {
	switch t {
	case datatypes.Unknown:
		return nil
	case datatypes.Any:
		return RandomValue(datatypes.Type((rand.Intn(datatypes.NUM_DATA_TYPES) + 1) - 2))
	case datatypes.String:
		return randString(rand.Intn(100))
	case datatypes.Float:
		return rand.Float32()
	case datatypes.Integer:
		return rand.Int()
	case datatypes.Boolean:
		return rand.Intn(10) > 4
	case datatypes.Object:
		return map[string]interface{}{}
	case datatypes.Array:
		return []interface{}{}
	case datatypes.Date:
		return time.Now().Add(time.Hour * 24 * time.Duration(rand.Intn(30)+1))
	case datatypes.Url:
		return &url.URL{
			Scheme: "http",
			Host:   "bit.ly",
			Path:   randString(6),
		}
	}

	return nil
}

// RandomStringValue is RandomValue that always returns a string
func RandomStringValue(t datatypes.Type) string {
	switch t {
	case datatypes.Unknown:
		return ""
	case datatypes.Any:
		return RandomStringValue(datatypes.Type((rand.Intn(datatypes.NUM_DATA_TYPES) + 1) - 2))
	case datatypes.String:
		return randString(rand.Intn(100))
	case datatypes.Float:
		str, _ := datatypes.Float.ValueToString(rand.Float32())
		return str
	case datatypes.Integer:
		str, _ := datatypes.Integer.ValueToString(rand.Int())
		return str
	case datatypes.Boolean:
		if rand.Intn(10) > 4 {
			return "true"
		} else {
			return "false"
		}
	case datatypes.Object:
		return "{}"
	case datatypes.Array:
		return "[]"
	case datatypes.Date:
		return time.Now().Add(time.Hour * 24 * time.Duration(rand.Intn(30)+1)).Format(time.ANSIC)
	case datatypes.Url:
		return "http://bit.ly/" + randString(6)
	}

	return ""
}