package refobj_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/gdey/gdjson/refobj"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func TestObjectUnmarshalJSON(t *testing.T) {

	type AuthStruct struct {
		Id   string             `json:"id"`
		Name refobj.Env[string] `json:"name"`
		Pass refobj.Env[string] `json:"pass"`
	}
	type FeaturesStruct struct {
		Name         refobj.Env[string]        `json:"name"`
		Auth         refobj.Object[AuthStruct] `json:"auth"`
		DisplayPrice refobj.Object[string]     `json:"display_price"`
	}

	type TestObject struct {
		Auths    []AuthStruct                   `json:"auths"`
		Features []FeaturesStruct               `json:"features"`
		Config   refobj.Object[json.RawMessage] `json:"config"`
	}
	const (
		data = `
{ "auths" : 
  [ { "id"   : "one"
		, "name" : "${env:USER}"
		, "pass" : "${env:EUID}"
    }
  , { "id"   : "two"
    , "name" : "boe"
    , "pass" : "456"
    }
  ]
, "features" :
	[ { "name" : "${env:USER}"
		, "auth" : "${path:auths[0]}"
    , "display_price": "$.10"
    }
  ]
, "config" : "${path:features[0]}"
}
`
	)
	var o TestObject
	err := refobj.UnmarshalJSON([]byte(data), &o)
	if err != nil {
		panic(err)
	}
	t.Log(spew.Sdump(o))
}
