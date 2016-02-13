package main

import (
    "bufio"
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "sort"
    "strconv"
    "strings"

    "github.com/alecthomas/kingpin"
    "github.com/gosuri/uitable"
    "github.com/xconstruct/vdir"

    "akeil.net/contacts"
)

// Add a new contact
func add(firstName string, lastName string, nickName string, skipEdit bool) error {
    var err error
    addressbook := contacts.NewAddressbook("/home/akeil/contacts")
    card := new(vdir.Card)
    card.Name.GivenName = []string{firstName}
    card.Name.FamilyName = []string{nickName}
    card.NickName = []string{nickName}

    if !skipEdit {
        //TODO err
        err = contacts.EditCard(card)
        if err != nil {
            // TODO: edit w/o change is an error
            return err
        }
    }
    err = contacts.Save(addressbook.Dirname, *card)
    if err != nil {
        return err
    }

    fmt.Println("Contact added.")
    return contacts.ShowDetails(*card)
}

// list all contacts matching the given `query`.
// Use an empty query to list all contacts.
func list(query string) error {
    addressbook := contacts.NewAddressbook("/home/akeil/contacts")
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
    sort.Sort(contacts.ByName(results))
    for _, card := range results {
        table.AddRow(contacts.FormatName(card),
                     contacts.PrimaryMail(card),
                     contacts.PrimaryPhone(card))
    }
    fmt.Println(table)
    return nil
}

// show details for a single contact that matches the given `query`.
// If multiple contacts match, user selects one.
func show(query string) error {
    addressbook := contacts.NewAddressbook("/home/akeil/contacts")
    card, err := selectOne(addressbook, query)
    if err != nil {
        return err
    }

    return contacts.ShowDetails(card)
}

// edit details for a single contact that matches the given `query`.
// If multiple contacts match, user selects one.
func edit(query string) error {
    addressbook := contacts.NewAddressbook("/home/akeil/contacts")
    card, err := selectOne(addressbook, query)
    if err != nil {
        return err
    }

    err = contacts.EditCard(&card)
    if err != nil {
        return err
    }
    // TODO check whether the card was changed
    err = contacts.Save("/home/akeil/contacts", card)
    if err != nil {
        return err
    }

    fmt.Println("Contact saved.")
    return contacts.ShowDetails(card)
}


// Helpers --------------------------------------------------------------------

func selectOne(book *contacts.Addressbook, query string) (vdir.Card, error) {
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


// Main -----------------------------------------------------------------------

func main() {
    verbose := kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()

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

    cmd := kingpin.Parse()
    if !*verbose {
        log.SetOutput(ioutil.Discard)
    }

    var err error
    switch cmd {
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