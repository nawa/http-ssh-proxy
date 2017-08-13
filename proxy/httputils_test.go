package proxy

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestReplaceAbsoluteLinks(t *testing.T) {
	s := `00000 <tag       href     ='/href1' attr="attr"> 11111 
	<tag2    href   =  "href2"> 222222 
	<tag3    src  =      "/src3"		/>   
	<tag4    src   =  'src4'>`

	expected := `00000 <tag       href     ='localhost:80/context/href1' attr="attr"> 11111 
	<tag2    href   =  "href2"> 222222 
	<tag3    src  =      "localhost:80/context/src3"		/>   
	<tag4    src   =  'src4'>`
	assert.Equal(t, expected, replaceAbsoluteLinks(s, "localhost:80/context"))
}

func TestReplaceLinks(t *testing.T) {
	s := `sdsds <tag href="http://prefix:9999/replaceFrom/path/to/something">`
	expected := `sdsds <tag href="http://prefix:9999/replaceTo/path/to/something">`
	replacements := []Replacement{
		{From: "replaceFrom", To: "replaceTo"},
	}
	assert.Equal(t, expected, replaceExternalLinks(s, replacements))
}
