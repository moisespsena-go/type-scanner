package scanner

import (
	"strings"
	"unicode"
)

type NameConverter interface {
	Convert(name string, next NameConverter) string
}

type NameConverterFunc func(name string, next NameConverter) string

func (NameConverterFunc) Convert(name string, next NameConverter) string {
	name = FieldName(name)
	if strings.HasSuffix(name, "Id") {
		name = strings.TrimSuffix(name, "Id") + "ID"
	}
	return name
}

var DefaultNameConverter NameConverter = NameConverterFunc(func(name string, next NameConverter) string {
	name = FieldName(name)
	if strings.HasSuffix(name, "Id") {
		name = strings.TrimSuffix(name, "Id") + "ID"
	}
	return name
})

var FakeNameConverter NameConverter = NameConverterFunc(func(name string, next NameConverter) string {
	return name
})

func FieldName(s string) string {
	var b []rune
	prev := '_'
	for _, r := range s {
		if r == '_' {
			prev = r
			continue
		}
		if prev == '_' {
			prev = r
			r = unicode.ToUpper(r)
		}
		b = append(b, r)
	}
	return string(b)
}
