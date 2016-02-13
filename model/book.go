package model

import (
    "errors"
    "fmt"
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
                //var card *vdir.Card
                card := new(vdir.Card)
                fullpath := ab.Dirname + "/" + file.Name()
                log.Printf("Load from %s", fullpath)
                data, err := ioutil.ReadFile(fullpath)
                if err != nil {
                    return err
                }
                // TODO panic if file does not end with empty an line
                err = vdir.Unmarshal(data, card)
                if err != nil {
                    return err
                }
                ab.cards = append(ab.cards, *card)
            }
        }
    }
    return nil
}

func matches(card vdir.Card, query string) bool {
    props := []string {
        strings.ToLower(card.FormattedName),
    }
    for _, s := range card.NickName {
        props = append(props, strings.ToLower(s))
    }
    for _, s := range card.Name.FamilyName {
        props = append(props, strings.ToLower(s))
    }
    for _, s := range card.Name.GivenName {
        props = append(props, strings.ToLower(s))
    }

    match := false
    for _, prop := range props {
        if prop != "" {
            if strings.HasSuffix(prop, query) || strings.HasPrefix(prop, query) {
                match = true
                break
            }
        }
    }
    return match
}

func Save(dirname string, card vdir.Card) error {
    uid := uuid.New()
    card.Version = "3.0"
    card.Uid = uid
    // rev, e.g. 1995-10-31T22:27:10Z
    card.Rev = time.Now().UTC().Format(time.RFC3339)
    card.FormattedName = FormatName(card)

    bytes, err := vdir.Marshal(card)
    fmt.Printf("Card: %s\n", bytes)
    return err
}

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
