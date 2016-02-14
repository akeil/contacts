package contacts

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pborman/uuid"
	"github.com/xconstruct/vdir"
)

type Addressbook struct {
	Dirname string
	cards   []vdir.Card
}

func NewAddressbook(dirname string) *Addressbook {
	book := new(Addressbook)
	book.Dirname = dirname
	return book
}

func (b *Addressbook) Find(query Query) ([]vdir.Card, error) {
	var err error
	var found []vdir.Card
	if b.cards == nil {
		err = b.load()
	}
	if err != nil {
		return found, err
	}

	for _, card := range b.cards {
		if query.Matches(card) {
			found = append(found, card)
		}
	}
	return found, err
}

// Save the given card
// the filename is derived from the cards UID.
// if no UID is set, one is assigned
// also set the Rev field
func (b Addressbook) Save(card vdir.Card) error {
	if card.Uid == "" {
		card.Uid = uuid.New()
	}
	// rev, e.g. 1995-10-31T22:27:10Z
	card.Rev = time.Now().UTC().Format(time.RFC3339)
	card.FormattedName = FormatName(card)

	bytes, err := vdir.Marshal(card)
	if err != nil {
		return err
	}

	fullpath := b.cardPath(card)
	file, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(bytes)
	return err
}

// delete the given card from the VDir
func (b Addressbook) Delete(card vdir.Card) error {
	path := b.cardPath(card)
	// TODO: move to trash
	return os.Remove(path)
}

func (b *Addressbook) load() error {
	log.Printf("Loading from %s", b.Dirname)
	b.cards = []vdir.Card{}

	info, err := os.Stat(b.Dirname)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("Not a directory")
	}
	dir, err := os.Open(b.Dirname)
	if err != nil {
		return err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".vcf" {
				card, err := loadCard(filepath.Join(b.Dirname, file.Name()))
				if err != nil {
					return err
				}
				b.cards = append(b.cards, *card)
			}
		}
	}
	return nil
}

func loadCard(fullpath string) (*vdir.Card, error) {
	card := new(vdir.Card)
	log.Printf("Load from %s", fullpath)
	data, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return card, err
	}
	// Unmarshal will panic if file does not end with empty an line
	// additional empty lines have no effect
	data = append(data, '\n')
	err = vdir.Unmarshal(data, card)
	if err != nil {
		return card, err
	}
	return card, nil
}

func (b Addressbook) cardPath(card vdir.Card) string {
	return filepath.Join(b.Dirname, card.Uid+".vcf")
}

// Sort Helper
type ByName []vdir.Card

func (a ByName) Len() int {
	return len(a)
}

func (a ByName) Swap(front, back int) {
	a[front], a[back] = a[back], a[front]
}

func (a ByName) Less(i, j int) bool {
	return FormatName(a[i]) < FormatName(a[j])
}

// search helper
func QueryTerm(term string) Query {
	return Query{term, []string{}}
}

type Query struct {
	Term       string
	Categories []string
}

func (q Query) Matches(card vdir.Card) bool {
	// Categories always match if not set
	categoryMatch := true
	if len(q.Categories) > 0 {
		categoryMatch = q.matchCategories(card)
	}
	termMatch := q.matchTerm(card)
	return categoryMatch && termMatch
}

func (q Query) matchCategories(card vdir.Card) bool {
	for _, requested := range q.Categories {
		for _, present := range card.Categories {
			if strings.ToLower(requested) == strings.ToLower(present) {
				return true
			}
		}
	}
	return false
}

func (q Query) matchTerm(card vdir.Card) bool {
	if contains(card.FormattedName, q.Term) {
		return true
	}
	if arrayContains(card.NickName, q.Term) {
		return true
	}
	if arrayContains(card.Name.FamilyName, q.Term) {
		return true
	}
	if arrayContains(card.Name.GivenName, q.Term) {
		return true
	}
	if typedValuesContain(card.Email, q.Term) {
		return true
	}
	if typedValuesContain(card.Telephones, q.Term) {
		return true
	}

	return false
}

func contains(s, q string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(q))
}

func typedValuesContain(tvalues []vdir.TypedValue, query string) bool {
	for _, tv := range tvalues {
		if contains(tv.Value, query) {
			return true
		}
	}
	return false
}

func arrayContains(texts []string, query string) bool {
	for _, s := range texts {
		if contains(s, query) {
			return true
		}
	}
	return false
}

// Create the Full Name from first name and last name
func FormatName(card vdir.Card) string {
	var name string
	if card.FormattedName != "" {
		name = card.FormattedName
	} else {
		if len(card.Name.GivenName) > 0 {
			name = card.Name.GivenName[0]
		}
		name = name + " "
		if len(card.Name.FamilyName) > 0 {
			name = name + card.Name.FamilyName[0]
		}
	}

	return strings.TrimSpace(name)
}

func PrimaryMail(card vdir.Card) string {
	var result string
	for _, mail := range card.Email {
		if mail.Value != "" {
			result = mail.Value
			break
		}
	}
	return result
}

func PrimaryPhone(card vdir.Card) string {
	var result string
	for _, phone := range card.Telephones {
		if phone.Value != "" {
			result = phone.Value
			break
		}
	}
	return result
}
