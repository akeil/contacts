package contacts

// go:generate go-bindata -pkg $GOPACKAGE -o assets.go tpl/

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"text/template"

	"github.com/gosuri/uitable"
	"github.com/xconstruct/vdir"
)

func FillTemplate(writer io.Writer, tplName string, card *vdir.Card) error {
	tpl, err := loadTemplate(tplName)
	if err != nil {
		return err
	}
	return tpl.Execute(writer, card)
}

func loadTemplate(name string) (*template.Template, error) {
	log.Println("Load template " + name)
	tpl := template.New(name)
	funcs := template.FuncMap{
		"join": join,
	}
	tpl.Funcs(funcs)

	// for `Asset` see https://github.com/jteeuwen/go-bindata
	tplStr, err := Asset("tpl/" + name)
	if err != nil {
		return tpl, err
	}
	return tpl.Parse(string(tplStr))
}

func join(list []string) string {
	return strings.Join(list, ", ")
}

// Render a list of cards
func ShowList(cards []vdir.Card, format string) {
	sort.Sort(ByName(cards))
	switch format {
	case "sup":
		renderSupContacts(cards)
	default:
		renderTable(cards)
	}
}

// render card data into a table for display
func renderTable(cards []vdir.Card) {
	table := uitable.New()
	table.Separator = "  "
	table.AddRow("NAME", "MAIL", "PHONE")

	for _, card := range cards {
		table.AddRow(FormatName(card),
			PrimaryMail(card),
			PrimaryPhone(card))
	}

	fmt.Println(table)
}

// render card data into a format suitable for sup contacts.txt.
// This is one line per contact, each line looks like this:
// {nick}: {firstName} {lastName} <{mail}>
func renderSupContacts(cards []vdir.Card) {
	template := "%v: %v <%v>\n"
	for _, card := range cards {
		// filter out cards that would create invalid entries
		if FormatNickName(card) == "" {
			continue
		} else if FormatName(card) == "" {
			continue
		} else if PrimaryMail(card) == "" {
			continue
		}

		fmt.Printf(template, strings.ToLower(FormatNickName(card)),
			FormatName(card), PrimaryMail(card))
	}
}
