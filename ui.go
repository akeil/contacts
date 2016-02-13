package contacts

import (
    "io"
    "log"
    "text/template"

    "github.com/xconstruct/vdir"
)

func FillTemplate(writer io.Writer, tplName string, card *vdir.Card) error {
    tpl, err := loadTemplate("/home/akeil/code/go/src/akeil.net/contacts",
                             tplName)
    if err != nil {
        return err
    }
    return tpl.Execute(writer, card)
}

func loadTemplate(basedir string, name string) (*template.Template, error) {
    fullpath := basedir + "/" + name
    log.Println("Load template " + fullpath)
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
