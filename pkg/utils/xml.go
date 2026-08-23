package utils

import (
	"bytes"
	"encoding/xml"
	"html"
	"regexp"
	"strings"
)

var (
	invalidXMLChars = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
)

func EscapeXML(s string) string {
	s = invalidXMLChars.ReplaceAllString(s, "")
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

func UnescapeXML(s string) string {
	return html.UnescapeString(s)
}

func EscapeCDATA(s string) string {
	s = invalidXMLChars.ReplaceAllString(s, "")
	return strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>")
}

func WrapCDATA(s string) string {
	return "<![CDATA[" + EscapeCDATA(s) + "]]>"
}

func SanitizeXML(s string) string {
	s = invalidXMLChars.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\x00", "")
	return s
}

func XMLHeader() string {
	return `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
}
