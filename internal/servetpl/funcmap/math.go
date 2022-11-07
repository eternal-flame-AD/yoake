package funcmap

import (
	"encoding/json"
	"fmt"

	"github.com/Knetic/govaluate"
)

func Math(expS string, args ...interface{}) (interface{}, error) {
	exp, err := govaluate.NewEvaluableExpressionWithFunctions(expS,
		map[string]govaluate.ExpressionFunction{
			"argv": func(arguments ...interface{}) (interface{}, error) {

				if len(arguments) != 1 {
					return nil, fmt.Errorf("argv expects 1 argument, got %d", len(arguments))
				}
				idx := int(arguments[0].(float64))

				if idx < 0 || idx > len(args) {
					return nil, fmt.Errorf("argv index out of range: %d", idx)
				}
				if idx == 0 {
					return expS, nil
				}

				vJ, err := json.Marshal(args[idx-1])
				if err != nil {
					return nil, err
				}
				var v interface{}
				if err := json.Unmarshal(vJ, &v); err != nil {
					return nil, err
				}

				return v, nil
			},
		})
	if err != nil {
		return nil, err
	}
	return exp.Evaluate(nil)
}
