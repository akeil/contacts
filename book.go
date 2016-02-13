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
    cards []vdir.Card
}

func NewAddressbook(dirname string) *Addressbook {
    book := new(Addressbook)
    book.Dirname = dirname
    return book
}

func (ab *Addressbook) Find(query string) ([]vdir.Card, error) {
    var err error
    var found []vdir.Card
    if ab.cards == nil {
        err = ab.load()
    }
    if err != nil {
        return found, err
    }

    for _, card := range ab.cards {
        if matches(card, query) {
            found = append(found, card)
        }
    }
    return found, err
}

func (ab *Addressbook) load() error{
    log.Printf("Loading from %s", ab.Dirname)
    ab.cards = []vdir.Card{}

    info, err := os.Stat(ab.Dirname)
    if err != nil {
        return err
    }
    if !info.IsDir() {
        return errors.New("Not a directory")
    }
    dir, err := os.Open(ab.Dirname)
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
                card, err := loadCard(ab.Dirname + "/" + file.Name())
                if err != nil {
                    return err
                }
                ab.cards = append(ab.cards, *card)
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
func matches(card vdir.Card, query string) bool {
    if contains(card.FormattedName, query){
        return true
    }
    if arrayContains(card.NickName, query) {
        return true
    }
    if arrayContains(card.Name.FamilyName, query) {
        return true
    }
    if arrayContains(card.Name.GivenName, query) {
        return true
    }
    if typedValuesContain(card.Email, query) {
        return true
    }
    if typedValuesContain(card.Telephones, query) {
        return true
    }

    return false
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

func contains(text, what string) bool {
    text, what = strings.ToLower(text), strings.ToLower(what)
    return strings.HasSuffix(text, what) || strings.HasPrefix(text, what)
}

// Save a vCard to the given directory
// the filename is derived from the cards UID.
// if no UID is set, one is assigned
// also sets the Rev field
func Save(dirname string, card vdir.Card) error {
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

    fullpath := dirname + "/" + card.Uid + ".vcf"
    file, err := os.Create(fullpath)
    if err != nil {
        return err
    }
    defer file.Close()
    _, err = file.Write(bytes)
    return err
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
