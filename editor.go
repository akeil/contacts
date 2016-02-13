package contacts

import (
    "bufio"
    "crypto/md5"
    "encoding/hex"
    "errors"
    "io"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "regexp"
    "strings"
    "github.com/xconstruct/vdir"
)

func EditCard(card *vdir.Card) error {
    tempfile, err := ioutil.TempFile("", "edit-card-")
    if err != nil {
        return err
    }

    defer os.Remove(tempfile.Name())
    err = FillTemplate(tempfile, "edit.tpl", card)
    if err != nil {
        return err
    }
    hashBefore := calcMd5(tempfile.Name())

    // run editor
    cmd := exec.Command("/usr/bin/gedit", tempfile.Name())
    err = cmd.Run()
    if err != nil {
        return err
    }

    if hashBefore != "" {
        hashAfter := calcMd5(tempfile.Name())
        log.Println("Hash Before: " + hashBefore)
        log.Println("Hash After:  " + hashAfter)
        if hashBefore == hashAfter {
            return errors.New("Contact was not changed")
        }
    }

    file, err := os.Open(tempfile.Name())
    if err != nil {
        return err
    }
    defer file.Close()

    reader := bufio.NewReader(file)
    scanner := bufio.NewScanner(reader)
    err = parseTemplate(scanner, card)
    if err != nil {
        return err
    }

    return err
}

func calcMd5(filename string) string {
    hash := md5.New()
    file, err := os.Open(filename)
    if err != nil {
        return ""
    }
    defer file.Close()
    if _, err := io.Copy(hash, file); err != nil {
        return ""
    }
    return hex.EncodeToString(hash.Sum(nil))
}

func parseTemplate(scanner *bufio.Scanner, card *vdir.Card) error {
    var line string
    f := parseNames
    for scanner.Scan() {
        line = scanner.Text()
        if strings.HasPrefix(line, "# Mail Adresses") {
            card.Email = []vdir.TypedValue{}
            f = parseMailAdress
        } else if strings.HasPrefix(line, "# Phone Numbers") {
            card.Telephones = []vdir.TypedValue{}
            f = parsePhoneNumber
        } else if strings.HasPrefix(line, "# Postal Addresses"){
            card.Addresses = []vdir.Address{}
            f = parsePostalAdress
        } else if strings.HasPrefix(line, "#") {
            continue
        } else if line == "" {
            continue
        }
        f(line, card)
    }

    return nil
}

var matchers = map[string] *regexp.Regexp {
    "firstName": regexp.MustCompile(`^First Name\s*: (.*?)$`),
    "lastName": regexp.MustCompile(`^Last Name\s*: (.*?)$`),
    "nickName": regexp.MustCompile(`^Nick\s*: (.*?)$`),
}

func parseNames(line string, card *vdir.Card) {
    for key, matcher := range matchers {
        if groups := matcher.FindStringSubmatch(line); groups != nil {
            value := strings.TrimSpace(groups[1])
            switch key {
            case "firstName":
                card.Name.GivenName = []string{value}
            case "lastName":
                card.Name.FamilyName = []string{value}
            case "nickName":
                card.NickName = []string{value}
            }
        }
    }
}

var typedValueRegex = regexp.MustCompile(`^([a-z]+)\s*:\s*(.*?)$`)

func parseMailAdress(line string, card *vdir.Card) {
    if groups := typedValueRegex.FindStringSubmatch(line); groups != nil {
        kind := groups[1]
        value := groups[2]
        tvalue := vdir.TypedValue{[]string{kind}, value}
        card.Email = append(card.Email, tvalue)
    }
}

func parsePhoneNumber(line string, card *vdir.Card) {
    if groups := typedValueRegex.FindStringSubmatch(line); groups != nil {
        kind := groups[1]
        value := groups[2]
        tvalue := vdir.TypedValue{[]string{kind}, value}
        card.Telephones = append(card.Telephones, tvalue)
    }
}

// Format "TYPE: ?; ?; STREET; CITY; REGION; POSTAL_CODE; COUNTRY"
var addrRegex = regexp.MustCompile(
    `^([a-z]+): (.*?); (.*?); (.*?); (.*?); (.*?); (.*?); (.*?)$`)
//                unk    unk    str    city   reg    code   country
func parsePostalAdress(line string, card *vdir.Card) {
    if groups := addrRegex.FindStringSubmatch(line); groups != nil {
        addr := vdir.Address{
            []string{strings.TrimSpace(groups[1])},
            "",  // Label
            "",  // PostOfficeBox
            "",  // ExtendedAddress
            strings.TrimSpace(groups[4]),  // Street
            strings.TrimSpace(groups[5]),  // Locality (City)
            strings.TrimSpace(groups[6]),  // Region
            strings.TrimSpace(groups[7]),  // PostalCode
            strings.TrimSpace(groups[8]),  // CountryName
        }
        card.Addresses = append(card.Addresses, addr)
    }
}

func ShowDetails(card vdir.Card) error {
    return FillTemplate(os.Stdout, "show.tpl", &card)
}
