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
    results := addressbook.Find(query)
    if len(results) == 0 {
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
    return editCard(&card)
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
    var err error
    found := book.Find(query)
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

var matchers = map[string] *regexp.Regexp {
    "firstName": regexp.MustCompile(`^First Name: (.*?)$`),
    "lastName": regexp.MustCompile(`^Last Name: (.*?)$`),
}

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

    cmd := exec.Command("/usr/bin/gedit", tempfile.Name())
    err = cmd.Run()
    if err != nil {
        return err
    }

    matches, err := parseTemplate(tempfile.Name())
    if err != nil {
        return err
    }

    card.Name.GivenName = []string{matches["firstName"]}
    //contact.LastName = matches["lastName"]
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

func parseTemplate(filename string) (map[string]string, error) {
    matches := map[string]string{}
    file, err := os.Open(filename)
    if err != nil {
        return matches, err
    }

    defer file.Close()
    reader := bufio.NewReader(file)
    scanner := bufio.NewScanner(reader)
    var line string
    for scanner.Scan() {
        line = scanner.Text()
        for key, matcher := range matchers {
            if groups := matcher.FindStringSubmatch(line); groups != nil {
                matches[key] = groups[1]
            }
        }
    }

    return matches, err
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
