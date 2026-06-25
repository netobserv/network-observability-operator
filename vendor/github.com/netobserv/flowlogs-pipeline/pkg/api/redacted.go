package api

import (
	"fmt"
	"log/slog"
)

type RedactedText string

func (RedactedText) Format(f fmt.State, _ rune) {
	_, _ = f.Write([]byte("[REDACTED]"))
}

func (RedactedText) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}
