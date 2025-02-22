package api

import "github.com/go-playground/validator/v10"

type validationResponse map[string][]string

func buildErrorsResponse(v validator.ValidationErrors) validationResponse {
	res := make(map[string][]string)

	for _, e := range v {
		if _, ok := res[e.Tag()]; !ok {
			res[e.Tag()] = make([]string, 0, 1)
		}

		res[e.Tag()] = append(res[e.Tag()], e.Error())
	}

	return res
}
