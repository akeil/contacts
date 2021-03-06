package contacts

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/xconstruct/vdir"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Start the configured editor with details of the given card.
// When the editor exits, apply changes to the card;
// return `true` if the card was modified.
func EditCard(cfg Configuration, card *vdir.Card) (bool, error) {
	modified := false
	tempfile, err := ioutil.TempFile("", "edit-card-")
	if err != nil {
		return modified, err
	}
	defer os.Remove(tempfile.Name())

	err = FillTemplate(tempfile, "edit.tpl", card)
	if err != nil {
		return modified, err
	}
	hashBefore := calcMd5(tempfile.Name())

	cmd := exec.Command(cfg.Editor, tempfile.Name())
	// see http://stackoverflow.com/questions/12088138/trying-to-launch-an-external-editor-from-within-a-go-program
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return modified, err
	}

	if hashBefore != "" {
		hashAfter := calcMd5(tempfile.Name())
		log.Println("Hash Before: " + hashBefore)
		log.Println("Hash After:  " + hashAfter)
		modified = hashBefore != hashAfter
	} else {
		// cannot compare hashes, assume modified
		modified = true
	}

	if !modified {
		return modified, nil
	}

	file, err := os.Open(tempfile.Name())
	if err != nil {
		return modified, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)
	err = parseTemplate(scanner, card)

	return modified, err
}

func calcMd5(path string) string {
	hash := md5.New()
	file, err := os.Open(path)
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

		} else if strings.HasPrefix(line, "# URLs") {
			card.Url = []vdir.TypedValue{}
			f = parseURL

		} else if strings.HasPrefix(line, "# Postal Addresses") {
			card.Addresses = []vdir.Address{}
			f = parsePostalAdress

		} else if strings.HasPrefix(line, "# Notes") {
			card.Note = ""
			f = parseNote
			continue

		} else if strings.HasPrefix(line, "#") {
			continue
		} else if line == "" {
			continue
		}
		f(line, card)
	}

	return nil
}

var matchers = map[string]*regexp.Regexp{
	"prefix":     regexp.MustCompile(`^Prefix\s*:(.*?)$`),
	"firstName":  regexp.MustCompile(`^First Name\s*:(.*?)$`),
	"lastName":   regexp.MustCompile(`^Last Name\s*:(.*?)$`),
	"nickName":   regexp.MustCompile(`^Nick\s*:(.*?)$`),
	"title":      regexp.MustCompile(`^Title\s*:(.*?)$`),
	"role":       regexp.MustCompile(`^Role\s*:(.*?)$`),
	"org":        regexp.MustCompile(`^Organization\s*:(.*?)$`),
	"categories": regexp.MustCompile(`^Categories\s*:(.*?)$`),
	"birthday":   regexp.MustCompile(`^Birthday\s*:(.*?)$`),
}

func parseNames(line string, card *vdir.Card) {
	for key, matcher := range matchers {
		if groups := matcher.FindStringSubmatch(line); groups != nil {
			value := strings.TrimSpace(groups[1])
			switch key {
			case "prefix":
				card.Name.HonorificNames = multiple(value)
			case "firstName":
				card.Name.GivenName = multiple(value)
			case "lastName":
				card.Name.FamilyName = multiple(value)
			case "nickName":
				card.NickName = multiple(value)
			case "title":
				card.Title = value
			case "role":
				card.Role = value
			case "org":
				card.Org = value
			case "categories":
				card.Categories = multiple(value)
			case "birthday":
				card.Birthday = value
			}
		}
	}
}

func multiple(value string) []string {
	values := strings.Split(value, ",")
	result := []string{}
	for _, value := range values {
		x := strings.TrimSpace(value)
		if x != "" {
			result = append(result, x)
		}
	}
	return result
}

func parseMailAdress(line string, card *vdir.Card) {
	value, err := typedValue(line)
	if err == nil {
		card.Email = append(card.Email, value)
	}
}

func parsePhoneNumber(line string, card *vdir.Card) {
	value, err := typedValue(line)
	if err == nil {
		card.Telephones = append(card.Telephones, value)
	}
}

func parseURL(line string, card *vdir.Card) {
	value, err := typedValue(line)
	if err == nil {
		card.Url = append(card.Url, value)
	}
}

func parseNote(line string, card *vdir.Card) {
	if card.Note != "" {
		card.Note += "\n"
	}
	card.Note += line
}

var typedValueRegex = regexp.MustCompile(`^([a-zA-Z][a-zA-Z, ]+?)\s*:\s*(.*?)$`)

func typedValue(line string) (vdir.TypedValue, error) {
	var result vdir.TypedValue
	var err error
	if groups := typedValueRegex.FindStringSubmatch(line); groups != nil {
		kinds := parseKinds(groups[1])
		value := groups[2]
		if value == "" {
			err = errors.New("")
		}
		result = vdir.TypedValue{kinds, value}
	} else {
		err = errors.New("")
	}
	return result, err
}

// Format "TYPE: ?; ?; STREET; CITY; REGION; POSTAL_CODE; COUNTRY"
var addrRegex = regexp.MustCompile(
	`^([a-z]+): (.*?); (.*?); (.*?); (.*?); (.*?); (.*?); (.*?)$`)

//                unk    unk    str    city   reg    code   country
func parsePostalAdress(line string, card *vdir.Card) {
	if groups := addrRegex.FindStringSubmatch(line); groups != nil {
		addr := vdir.Address{
			parseKinds(groups[1]), //  Types
			"", // Label
			"", // PostOfficeBox
			"", // ExtendedAddress
			strings.TrimSpace(groups[4]), // Street
			strings.TrimSpace(groups[5]), // Locality (City)
			strings.TrimSpace(groups[6]), // Region
			strings.TrimSpace(groups[7]), // PostalCode
			strings.TrimSpace(groups[8]), // CountryName
		}
		card.Addresses = append(card.Addresses, addr)
	}
}

func parseKinds(kindstr string) []string {
	kinds := strings.Split(kindstr, ",")
	for index := range kinds {
		kinds[index] = strings.TrimSpace(strings.ToLower(kinds[index]))
	}
	return kinds
}

func ShowDetails(card vdir.Card) error {
	return FillTemplate(os.Stdout, "show.tpl", &card)
}
