package args

import (
	"net/url"
	"strings"
)

// ParseInitialArg inspects the argument list and returns either a file path to
// open or a hostname extracted from a kpass://fill?host= URI.
func ParseInitialArg(args []string) (filePath, fillHost string) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if strings.HasPrefix(arg, "kpass://") {
			u, err := url.Parse(arg)
			if err == nil && u.Host == "fill" {
				fillHost = u.Query().Get("host")
			}
			return "", fillHost
		}
		return arg, ""
	}
	return "", ""
}
