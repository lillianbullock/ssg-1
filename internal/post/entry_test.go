package post

import (
	"io/ioutil"
	// "fmt"
	// "os"
	// "reflect"
	"testing"
)

func TestStripPageConfig(t *testing.T) {
	type test struct {
		input string
		want  string
	}

	tests := []test{
		{input: `<!-- CONFIG
frog
CONFIG --> 
<p>Hello</p>`,
			want: `<p>Hello</p>`,
		},
		{input: `<p>Hello</p>`,
			want: `<p>Hello</p>`,
		},
		{input: `<!-- comment --> 
		<p>Hello</p>`,
			want: `<!-- comment --> 
		<p>Hello</p>`,
		},
	}

	for _, tc := range tests {
		output := string(stripPageConfig([]byte(tc.input)))
		if tc.want != output {
			t.Errorf("expected: %v, got: %v", tc.want, output)
		}
	}
}

func TestGetPageConfig(t *testing.T) {
	config := `<!-- CONFIG
title: Title
snippet: Snippet about the article
date: 2020-03-09T22:00:00Z
CONFIG -->

<h2 id="subtitle-1">Subtitle 1</h2>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua.</p>

<h2 id="subtitle-2">Subtitle 2</h2>
<p>Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip
ex ea commodo consequat.</p>
<ul>
<li>one</li>
<li>Two</li>
</ul>

<p class="sig">Creator</p>
<p class="date">Published on March 9, 2020</p>`
	
	fileName := "tempConfig"

	err := ioutil.WriteFile(fileName, []byte(config), 0644) 
	if err != nil{
		t.Errorf("error writing file: " + err.Error())
		return
	}

	sitePath := "sitepath"
	// TODO pass in a SiteInfo with non-default data
	entry, err := GetPageConfig(fileName, sitePath, SiteInfo{})
	if err != nil{
		t.Errorf("error reading file: " + err.Error())
		return
	}

	if entry.FilePath != fileName {
		t.Errorf("filepath should be: %s, got: %s", fileName, entry.FilePath)
	}
	if entry.SitePath != sitePath {
		t.Errorf("sitepath should be: %s, got: %s", sitePath, entry.SitePath)
	}
	if string(entry.Content) != string(stripPageConfig([]byte(config))) { 
		// is this the best way to do this one?
		t.Errorf("html parsed improperly")	
	}

	//TODO compare more parts of the struct
	
}

// type Entry struct {
// 	FilePath     string
// 	SitePath     string
// 	Type         string   `yaml:"type"`
// 	Title        string   `yaml:"title"`
// 	Snippet      string   `yaml:"snippet"`
// 	Image        string   `yaml:"image"`
// 	Date         string   `yaml:"date"`
// 	RelatedPosts []string `yaml:"related"`
// 	Author       string   `yaml:"author"`
// 	Content      []byte
// 	HTML         []byte
// }


// type SiteInfo struct {
// 	DefaultTitle string
// 	DefaultImage string
// }