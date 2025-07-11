package refobj

import (
	"encoding/json"
	"fmt"
	"log"
)

// unmarshal will attempt to determine the type of unmarhsalling that need to occure.
// if data is of type string, it will process that string and send it to matched (along with if final is of type string) to see if special processing of the field is needed. If matched returns an error, and the final object ~= string, then the data will be unmarshalled into final, and the unmarshalling error will be returned. Otherwise matched's error will be returned.
func unmarshal[T any](data []byte, final *T, matched func(m string, objIsString bool) error) (hydrated bool, err error) {
	var (
		obj T
		str string
	)
	_, objIsString := any(obj).(string)
	log.Printf("data: %s", data)
	log.Printf("objIsString: %v\n", objIsString)

	err = json.Unmarshal(data, &str)
	if err != nil {
		// well it's not a string, so just Unmarsal regularly and return the error
		err = json.Unmarshal(data, final)
		if err != nil {
			return true, fmt.Errorf("unmarshal data(%s) to final: %w", data, err)
		}
		return true, nil
	}

	// We have a string, it could be a value we are interested in,
	// or if ObjeIsString is true then it's value
	err = matched(str, objIsString)
	if err != nil && objIsString {
		log.Printf("string data: %s, %v", data, final)
		return true, json.Unmarshal(data, final)
	}
	return false, err
}
