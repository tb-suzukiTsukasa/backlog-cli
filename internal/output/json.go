package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// PrintJSON は v を JSON 形式で w に出力する。
func PrintJSON(w io.Writer, v any) error {
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("JSON 出力エラー: %w", err)
	}
	return nil
}
