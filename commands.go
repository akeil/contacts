package main

import (
    "bufio"
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
    "text/template"

    "github.com/alecthomas/kingpin"
    "github.com/gosuri/uitable"
    "github.com/xconstruct/vdir"

    "akeil.net/contacts/model"
)

// Add a new contact
func add(firstName string, lastName string, nickName string, skipEdit bool) error{
    addressbook := model.NewAddressbook("/home/akeil/contacts")
    card := new(vdir.Card)
    card.Name.GivenName = []string{firstName}
    card.Name.FamilyName = []string{nickName}
    card.NickName = []string{nickName}

    if !skipEdit {
        //TODO err
        editCard(card)
    }
    model.Save(addressbook.Dirname, *card)
    return nil
}

// list all contacts matching the given `query`.
// Use an empty query to list all contacts.
func list(query string) error {
    addressbook := model.NewAddressbook("/home/akeil/contacts")
    results, err := addressbook.Find(query)
    if err != nil {
        return err
    } else if len(results) == 0 {
        fmt.Println("No match.")
        return nil
    }

    table := uitable.New()
    table.Separator = "  "
    table.AddRow("NAME", "MAIL", "PHONE")
    for _, card := range results {
        table.AddRow(model.FormatName(card),
                     model.PrimaryMail(card),
                     model.PrimaryPhone(card))
    }
    fmt.Println(table)
    return nil
}

// show details for a single contact that matches the given `query`.
// If multiple contacts match, user selects one.
func show(query string) error {
    addressbook := model.NewAddressbook("/home/akeil/contacts")
    card, err := selectOne(addressbook, query)
    if err != nil {
        return err
    }

    return showDetails(card)
}

// edit details for a single contact that matches the given `query`.
// If multiple contacts match, user selects one.
func edit(query string) error {
    addressbook := model.NewAddressbook("/home/akeil/contacts")
    card, err := selectOne(addressbook, query)
    if err != nil {
        return err
    }

    err = editCard(&card)
    if err != nil {
        return err
    }
    // TODO check whether the card was changed
    return model.Save("/home/akeil/contacts", card)
}


// Helpers --------------------------------------------------------------------

func showDetails(card vdir.Card) error {
    fmt.Println(card.Email)
    tpl, err := loadTemplate("/home/akeil/code/go/src/akeil.net/contacts",
                             "details.tpl")
    if err != nil {
        return err
    }
    return tpl.Execute(os.Stdout, card)
}

func loadTemplate(basedir string, name string) (*template.Template, error) {
    fullpath := basedir + "/" + name
    tpl := template.New(name)

    funcs := template.FuncMap{
        "join": join,
    }
    tpl.Funcs(funcs)

    tpl, err := tpl.ParseFiles(fullpath)
    return tpl, err
}

func join(list []string) string {
    result := ""
    for i := 0; i < len(list); i++ {
        if i > 0 {
            result += ", "
        }
        result += list[i]
    }
    return result
}

func selectOne(book *model.Addressbook, query string) (vdir.Card, error) {
    var selected vdir.Card
    found, err := book.Find(query)
    if err != nil {
        return selected, err
    }

    if len(found) > 1 {
        selected, err = choose(found)
    } else if len(found) == 1 {
        selected = found[0]
    } else {
        err = errors.New("No match.")
    }
    return selected, err
}

func choose(choices []vdir.Card) (vdir.Card, error) {
    var chosen vdir.Card
    fmt.Println("Select a contact:")
    for i :=0; i < len(choices); i++ {
        fmt.Print(i + 1)
        fmt.Print(") ")
        fmt.Println(displayName(choices[i]))
    }
    fmt.Print("> ")
    console := bufio.NewReader(os.Stdin)
    input, err := console.ReadString('\n')
    if err != nil {
        return chosen, err
    }

    index, err := strconv.ParseInt(strings.TrimSpace(input), 10, 0)
    if err != nil {
        return chosen, err
    }

    // TODO check index > 0 and <= len
    index -= 1
    chosen = choices[index]
    return chosen, err
}

func displayName(card vdir.Card) string {
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


// Editor ---------------------------------------------------------------------

func editCard(card *vdir.Card) error {
    tempfile, err := ioutil.TempFile("", "edit-card-")
    if err != nil {
        return err
    }

    defer os.Remove(tempfile.Name())
    err = fillTemplate(tempfile, card)
    if err != nil {
        return err
    }

    // run editor
    cmd := exec.Command("/usr/bin/gedit", tempfile.Name())
    err = cmd.Run()
    if err != nil {
        return err
    }

    // TODO check for change
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

func fillTemplate(file *os.File, card *vdir.Card) error {
    tpl, err := loadTemplate("/home/akeil/code/go/src/akeil.net/contacts",
                             "edit.tpl")
    if err != nil {
        return err
    }
    return tpl.Execute(file, card)
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


// Main -----------------------------------------------------------------------

func main() {
    addCmd := kingpin.Command("add", "Add a contact.")
    addFirstName := addCmd.Flag("first", "First Name").Short('f').String()
    addLastName := addCmd.Flag("last", "Last Name").Short('l').String()
    addNick := addCmd.Flag("nick", "Nick Name").Short('n').String()
    addSkipEdit := addCmd.Flag("no-edit", "Skip editor").Short('E').Bool()

    listCmd := kingpin.Command("list", "List contacts.")
    listQuery := listCmd.Arg("query", "Search term.").String()

    editCmd := kingpin.Command("edit", "Edit a contact.")
    editQuery := editCmd.Arg("query", "Search term.").String()
    showCmd := kingpin.Command("show", "Show contact details.")
    showQuery := showCmd.Arg("query", "Search term.").String()

    //log.SetOutput(ioutil.Discard)
    var err error
    log.Println("start")
    switch kingpin.Parse() {
    case "add":
        err = add(*addFirstName, *addLastName, *addNick, *addSkipEdit)
    case "list":
        err = list(*listQuery)
    case "show":
        err = show(*showQuery)
    case "edit":
        err = edit(*editQuery)
    }

    if err != nil {
        fmt.Println(err)
    }
}
