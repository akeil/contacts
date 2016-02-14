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


// Controller -----------------------------------------------------------------

type controller struct {
    term string
    categories string
    firstName string
    lastName string
    nickName string
    skipEdit bool
}

func (c *controller) query() contacts.Query {
    return contacts.Query{c.term, normalizedSplit(c.categories)}
}

func normalizedSplit(s string) []string {
    result := []string{}
    parts := strings.Split(s, ",")
    for _, part := range parts {
        x := strings.TrimSpace(part)
        if x != "" {
            result = append(result, part)
        }
    }
    return result
}

func (c *controller) card() vdir.Card {
    card := vdir.Card{}
    card.Name.GivenName = []string{c.firstName}
    card.Name.FamilyName = []string{c.lastName}
    card.NickName = []string{c.nickName}
    return card
}

// Add a new contact
func (c *controller) add(unused *kingpin.ParseContext) error {
    var err error
    cfg := contacts.ReadConfiguration()
    book := contacts.NewAddressbook(cfg.Addressbook)
    card := c.card()
    if !c.skipEdit {
        err = contacts.EditCard(cfg, &card)
        if err != nil {
            // TODO: edit w/o change is an error
            return err
        }
    }
    err = book.Save(card)
    if err != nil {
        return err
    }

    fmt.Println("Contact added.")
    return contacts.ShowDetails(card)
}

// list all contacts matching the given `query`.
// Use an empty query to list all contacts.
func (c *controller) list(unused *kingpin.ParseContext) error {
    cfg := contacts.ReadConfiguration()
    book := contacts.NewAddressbook(cfg.Addressbook)
    results, err := book.Find(c.query())
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
func (c *controller) show(unused *kingpin.ParseContext) error {
    cfg := contacts.ReadConfiguration()
    book := contacts.NewAddressbook(cfg.Addressbook)
    card, err := selectOne(book, c.query())
    if err != nil {
        return err
    }
    return contacts.ShowDetails(card)
}

// edit details for a single contact that matches the given `query`.
// If multiple contacts match, user selects one.
func (c *controller) edit(unused *kingpin.ParseContext) error {
    cfg := contacts.ReadConfiguration()
    book := contacts.NewAddressbook(cfg.Addressbook)
    card, err := selectOne(book, c.query())
    if err != nil {
        return err
    }

    err = contacts.EditCard(cfg, &card)
    if err != nil {
        return err
    }
    err = book.Save(card)
    if err != nil {
        return err
    }

    fmt.Println("Contact saved.")
    return contacts.ShowDetails(card)
}

// delete a contact
func (c *controller) del(unused *kingpin.ParseContext) error {
    cfg := contacts.ReadConfiguration()
    book := contacts.NewAddressbook(cfg.Addressbook)
    card, err := selectOne(book, c.query())
    if err != nil {
        return err
    }

    err = book.Delete(card)
    if err == nil {
        fmt.Println("Contact deleted.")
    }
    return err
}


// Helpers --------------------------------------------------------------------

func selectOne(book *contacts.Addressbook, query contacts.Query) (vdir.Card, error) {
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

func catFlag(cmd *kingpin.CmdClause, ctl *controller) {
    cmd.Flag("categories", "Categories to search, comma separated.").
        Short('c').
        StringVar(&ctl.categories)
}

func queryArg(cmd *kingpin.CmdClause, ctl *controller) {
    cmd.Arg("query", "Search term.").StringVar(&ctl.term)
}

var verbose bool

func verbosity(unused *kingpin.ParseContext) error {
    if !verbose {
        log.SetOutput(ioutil.Discard)
    }
    return nil
}

func main() {
    ctl := &controller{}

    app := kingpin.New("contacts", "Manage contacts in a VDir.")
    app.HelpFlag.Short('h')  // -h for --help
    app.Flag("verbose", "Verbose mode.").Short('v').BoolVar(&verbose)
    app.Action(verbosity)

    ls := app.Command("ls", "List contacts").Action(ctl.list)
    catFlag(ls, ctl)
    queryArg(ls, ctl)

    add := app.Command("add", "Add a new contact.").Action(ctl.add)
    add.Flag("first", "First Name").Short('f').StringVar(&ctl.firstName)
    add.Flag("last", "Last Name").Short('l').StringVar(&ctl.lastName)
    add.Flag("nick", "Nick Name").Short('n').StringVar(&ctl.nickName)
    add.Flag("no-edit", "Skip editor").Short('E').BoolVar(&ctl.skipEdit)

    show := app.Command("show", "Show contact details.").Action(ctl.show)
    catFlag(show, ctl)
    queryArg(show, ctl)

    edit := app.Command("edit", "Edit contacts.").Action(ctl.edit)
    catFlag(edit, ctl)
    queryArg(edit, ctl)

    del := app.Command("del", "Delete a contact.").Action(ctl.del)
    catFlag(del, ctl)
    queryArg(del, ctl)

    kingpin.MustParse(app.Parse(os.Args[1:]))
}
