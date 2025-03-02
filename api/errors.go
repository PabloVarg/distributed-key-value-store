package api

import "github.com/go-playground/validator/v10"

type validationResponse map[string][]string

func buildErrorsResponse(v validator.ValidationErrors) validationResponse {
	res := make(map[string][]string)

	for _, e := range v {
		if _, ok := res[e.Field()]; !ok {
			res[e.Field()] = make([]string, 0, 1)
		}

		res[e.Field()] = append(res[e.Field()], e.Tag())
	}

	return res
}
