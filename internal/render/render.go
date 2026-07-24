package render

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Output prints v as indented JSON when jsonOut is true; otherwise it calls
// table to render the human view to stdout.
func Output(jsonOut bool, v any, table func(w io.Writer)) error {
	if jsonOut {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}
	table(os.Stdout)
	return nil
}
