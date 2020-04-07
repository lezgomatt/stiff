package main

import "strings"

type AcceptEncoding struct {
	Brotli bool
	GZip   bool
}

func parseAcceptEncoding(headerText string) AcceptEncoding {
	var ae AcceptEncoding

	for _, part := range strings.Split(headerText, ",") {
		var enc string
		if sc := strings.Index(part, ";"); sc != -1 {
			// ignore quality values, we always prioritize brotli over gzip
			enc = strings.TrimSpace(part[:sc])
		} else {
			enc = strings.TrimSpace(part)
		}

		switch enc {
		case "br":
			ae.Brotli = true
		case "gzip":
			ae.GZip = true
		}
	}

	return ae
}
