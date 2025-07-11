package refobj_test

import (
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
		Id   string `json:"id"`
		Name string `json:"name"`
		Pass string `json:"pass"`
	}
	type FeaturesStruct struct {
		Name         refobj.Env[string]        `json:"name"`
		Auth         refobj.Object[AuthStruct] `json:"auth"`
		DisplayPrice refobj.Object[string]     `json:"display_price"`
	}

	type TestObject struct {
		Auths    []AuthStruct     `json:"auths"`
		Features []FeaturesStruct `json:"features"`
	}
	const (
		data = `
{ "auths" : 
  [ { "id"   : "one"
    , "name" : "joe"
    , "pass" : "123"
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
