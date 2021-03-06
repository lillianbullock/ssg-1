package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"

	"github.com/shortmoose/ssg/internal/config"
	"github.com/shortmoose/ssg/internal/post"
	"github.com/shortmoose/ssg/internal/util"
)

var (
	cfg config.Config
)

// feed TODO
type feed struct {
	SiteURL   string
	SiteTitle string
	SiteID    string
	Author    string
}

func createAtomFeed(path string, feed feed, configs []post.Entry) error {
	ents := []post.Entry{}
	for i := range configs {
		if configs[i].Date != "" {
			ents = append(ents, configs[i])
		}
	}
	if len(ents) == 0 {
		return fmt.Errorf("Can't create XML feed, no entries")
	}
	sort.Sort(post.ByDate(ents))

	s := ""
	s += fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	s += fmt.Sprintf("<feed xmlns=\"http://www.w3.org/2005/Atom\">\n")
	s += fmt.Sprintf("  <title>%s</title>\n", feed.SiteTitle)
	s += fmt.Sprintf("  <link href=\"%s/\" />\n", feed.SiteURL)
	s += fmt.Sprintf("  <updated>%s</updated>\n", ents[0].Date)
	s += fmt.Sprintf("  <id>%s</id>\n", feed.SiteID)

	for _, e := range ents {
		if e.Date != "" {
			s += fmt.Sprintf("<entry>\n")
			s += fmt.Sprintf("  <title>%s</title>\n", e.Title)
			s += fmt.Sprintf("  <link href=\"%s%s\" />\n", feed.SiteURL, e.SitePath)
			s += fmt.Sprintf("  <updated>%s</updated>\n", e.Date)
			s += fmt.Sprintf("  <id>%s%s</id>\n", feed.SiteURL, e.SitePath)
			s += fmt.Sprintf("  <author><name>%s</name></author>\n", feed.Author)
			s += fmt.Sprintf("  <content type=\"html\"><![CDATA[\n")
			s += fmt.Sprintf("%s\n", e.Content)
			s += fmt.Sprintf("  ]]></content>\n")
			s += fmt.Sprintf("</entry>\n")
		}
	}
	s += fmt.Sprintf("</feed>\n")

	body := []byte(s)
	body = bytes.ReplaceAll(body, []byte("/img/"), []byte(cfg.ImageURL+"/"))
	body = bytes.ReplaceAll(body, []byte("/pdf/"), []byte(cfg.ImageURL+"/"))
	body = bytes.ReplaceAll(body, []byte("href=\"/"), []byte("href=\""+cfg.URL+"/"))

	err := ioutil.WriteFile(path, body, 0644)
	if err != nil {
		return fmt.Errorf("WriteFile :%w", err)
	}

	return nil
}

func postIndexEntry(e post.Entry) string {
	var cnt string
	img := e.Image
	if img == "" {
		img = cfg.Image
	}
	cnt += "\n\n"
	cnt += "<div style=\"width: 100%; overflow: hidden;\">"
	cnt += "<div style=\"width: 170px; float: left;\">"
	cnt += "<a href=\"" + e.SitePath + "\">"
	cnt += "<img class=\"himg\" alt=\"thumbnail\" src=\"" + img + "\"></a></div>"
	cnt += "<div style=\"margin-left: 190px;\">"

	cnt += "<a href=\"" + e.SitePath + "\"><b>" + e.Title + "</b></a>"
	if e.Snippet != "" {
		cnt += "<p>" + e.Snippet + "</p>"
	}
	cnt += "</div></div><br />\n"

	return cnt
}

func postIndexEntryKey(key string, configs []post.Entry) (string, error) {
	for i := range configs {
		if configs[i].SitePath == key {
			return postIndexEntry(configs[i]), nil
		}
	}

	return "", fmt.Errorf("invalid key: ''%s'", key)
}

func buildIndex(path string, ent post.Entry, configs []post.Entry) error {
	ents := []post.Entry{}
	for i := range configs {
		if configs[i].Date != "" {
			ents = append(ents, configs[i])
		}
	}
	if len(ents) == 0 {
		return fmt.Errorf("Can't create XML feed, no entries")
	}
	sort.Sort(post.ByDate(ents))

	cnt := ""
	for _, e := range ents {
		cnt += postIndexEntry(e)
	}

	ent.Content = []byte(cnt)
	err := buildPage(path, ent, configs)
	if err != nil {
		return err
	}

	return nil
}

func buildPage(dest string, ent post.Entry, configs []post.Entry) error {
	pre, err := ioutil.ReadFile("templates/pre.html")
	if err != nil {
		return err
	}

	pre = bytes.Replace(pre, []byte("<!--TITLE-->"), []byte(ent.Title), -1)
	pre = bytes.Replace(pre, []byte("<!--DESCRIPTION-->"), []byte(ent.Snippet), -1)
	pre = bytes.Replace(pre, []byte("<!--IMAGE-->"), []byte(ent.Image), -1)

	meta := ""
	if ent.Image != cfg.Image {
		meta = "<meta property=\"og:image\" content=\"" + ent.Image + "\" />\n  "
	}
	pre = bytes.Replace(pre, []byte("<!--META-->\n"), []byte(meta), -1)

	if ent.Title != cfg.Title {
		pre = append(pre, []byte("<h1>"+ent.Title+"</h1>\n")...)
	}

	body := ent.Content

	var errStrings []string
	re := regexp.MustCompile(`<!--/.*-->`)
	body = re.ReplaceAllFunc(body, func(a []byte) []byte {
		key := string(a[4 : len(a)-3])
		html, err := postIndexEntryKey(key, configs)
		if err != nil {
			errStrings = append(errStrings, key)
			return []byte("")
		}
		return []byte(html)
	})
	if len(errStrings) != 0 {
		return fmt.Errorf("Invalid keys: %v", errStrings)
	}

	var extra string
	for _, k := range ent.RelatedPosts {
		html, err := postIndexEntryKey(k, configs)
		if err != nil {
			return err
		}
		extra += html
	}
	if len(extra) != 0 {
		extra = "\n<hr class=\"foo\">\n" +
			"<p><b>If you enjoyed that article, try out a couple more:</b></p>\n" +
			extra + "\n\n"
	}
	body = append(body, []byte(extra)...)

	post, err := ioutil.ReadFile("templates/post.html")
	if err != nil {
		return fmt.Errorf("ReadFile :%w", err)
	}
	body = append(pre, append(body, post...)...)

	body = bytes.ReplaceAll(body, []byte("/img/"), []byte(cfg.ImageURL+"/"))
	body = bytes.ReplaceAll(body, []byte("/pdf/"), []byte(cfg.ImageURL+"/"))

	err = ioutil.WriteFile(dest, body, 0644)
	if err != nil {
		return fmt.Errorf("WriteFile :%w", err)
	}

	return nil
}

func validateImagesExist(configs []post.Entry) error {
	m := map[string]bool{}
	for _, ent := range configs {
		re := regexp.MustCompile(`/(img|pdf)/[^"']*`)
		for _, url := range re.FindAll([]byte(string(ent.Content)+" "+ent.Image), -1) {
			urlstr := cfg.ImageURL + string(url)[4:]
			m[urlstr] = true
		}
	}

	for s := range m {
		res, err := http.Head(s)
		if err != nil {
			fmt.Printf("Error while looking for: %s\n", s)
			return err
		} else if res.StatusCode != 200 {
			fmt.Printf("Error: %s returned status %d\n", s, res.StatusCode)
			return fmt.Errorf("Blah")
		}
	}

	return nil
}

func walk() error {
	var siteinfo post.SiteInfo
	siteinfo.DefaultTitle = cfg.Title
	siteinfo.DefaultImage = cfg.Image

	var configs []post.Entry
	err := util.Walk("posts", func(path string, info os.FileInfo) error {
		ent, err := post.GetPageConfig(path, path[5:], siteinfo)
		if err != nil {
			return err
		}

		configs = append(configs, ent)
		return nil
	})
	if err != nil {
		return err
	}

	for _, ent := range configs {
		if ent.Type == "index" {
			err := buildIndex("website/posts"+ent.SitePath, ent, configs)
			if err != nil {
				return fmt.Errorf("parsing %s: %w", ent.FilePath, err)
			}
		} else {
			err = buildPage("website/posts"+ent.SitePath, ent, configs)
			if err != nil {
				return fmt.Errorf("parsing %s: %w", ent.FilePath, err)
			}
		}
	}

	err = validateImagesExist(configs)
	if err != nil {
		return err
	}

	var feed feed
	feed.SiteTitle = cfg.Title
	feed.SiteURL = cfg.URL
	feed.SiteID = cfg.URL + "/"
	feed.Author = cfg.Author

	err = createAtomFeed("website/atom.xml", feed, configs)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	cfgTmp, err := config.GetConfig("ssg.yaml")
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	cfg = cfgTmp

	err = walk()
	if err != nil {
		log.Fatalf("%v\n", err)
	}
}
