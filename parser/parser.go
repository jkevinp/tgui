package parser

import (
	"fmt"
	"reflect"
	"strings"
)

func ParseTGTags(v interface{}) (map[string]map[string]string, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		fmt.Println("Provided value is not a struct")
		return nil, fmt.Errorf("provided value is not a struct")
	}

	result := make(map[string]map[string]string)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("tg")
		// fmt.Printf("Field: %s, tg Tag: %s\n", field.Name, tag)
		tags := strings.Split(tag, ";")
		key := field.Name
		tagsMap := make(map[string]string)
		for _, t := range tags {

			parts := strings.Split(t, ":")

			if len(parts) == 2 {
				// fmt.Printf("    %s: %s\n", parts[0], parts[1])
				tagsMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			} else {
				// fmt.Printf("    %s\n", parts[0])
				if t != "" {
					tagsMap[t] = "true"
				}

			}
		}

		result[key] = tagsMap

	}

	return result, nil
}
