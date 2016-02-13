package contacts

import (
//    "errors"
//    "io"
//    "log"
    "text/template"

//    "github.com/xconstruct/vdir"
)

func LoadTemplate(basedir string, name string) (*template.Template, error) {
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
