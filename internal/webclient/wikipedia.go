package webclient

import (
	"io"

	"github.com/PuerkitoBio/goquery"
)

type Cleaner struct {
	BodySelector string
	RemoveList   []string
}

var WikipediaCleaner = Cleaner{
	BodySelector: `[id="bodyContent"]`,
	RemoveList: []string{
		"* > a[href^=\"/w/index.php\"]",
		"form[action^=\"/w/index.php\"]",
		"[id^=\"cite_note-\"]",
		"[id^=\"cite_ref-\"]",
		"[id^=\"footer-places\"]",
		"[id=References]",
		"[id=centralNotice]",
		"[id=catlinks]",
	},
}

func (c *Cleaner) Parse(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", err
	}
	docc := doc.Clone()
	if c.BodySelector != "" {
		docc = doc.Find(c.BodySelector)
	}

	for _, q := range c.RemoveList {
		docc.Find(q).Remove()
	}

	return docc.Html()
}
