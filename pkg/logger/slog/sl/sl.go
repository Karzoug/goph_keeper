package sl

import (
	"golang.org/x/exp/slog"
)

func Error(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
