package contacts

// go:generate go-bindata -pkg $GOPACKAGE -o assets.go tpl/

import (
    "io"
    "log"
    "text/template"

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
    result := ""
    for i := 0; i < len(list); i++ {
        if i > 0 {
            result += ", "
        }
        result += list[i]
    }
    return result
}
