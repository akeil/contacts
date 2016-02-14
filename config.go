package contacts

import (
    "encoding/json"
    "io"
    "log"
    "os"
    "os/user"
    "path/filepath"
    "strings"
)


type Configuration struct {
    Addressbook string
    Editor string
}

func ReadConfiguration() Configuration {
    var config Configuration
    // for `Asset` see https://github.com/jteeuwen/go-bindata
    defaultConf, err := Asset("config/default.json")
    if err != nil {
        log.Println("Could not read default conf.")
        log.Println(err)
    }
    err = json.Unmarshal(defaultConf, &config)

    paths := []string{
        "/etc/contacts.config.json",
        "/home/akeil/.config/contacts.config.json",
    }
    for _, path := range paths {
        file, err := os.Open(path)
        if err != nil {
            log.Println("Could not read " + path)
            log.Println(err)
            continue
        }
        decoder := json.NewDecoder(file)
        //cfg := Configuration{"", ""}
        for {
            if err := decoder.Decode(&config); err == io.EOF {
                break
            } else if err != nil {
                log.Println("Could not read " + path)
                log.Println(err)
                break
            }
        }
    }

    config.Addressbook = replaceHomeDir(config.Addressbook)
    config.Editor = replaceHomeDir(config.Editor)
    return config
}

func replaceHomeDir(path string) string {
    user, err := user.Current()
    if err != nil {
        return path
    }

    if strings.HasPrefix(path, "~") {
        return filepath.Join(user.HomeDir, path[1:])
    }else if strings.HasPrefix(path, "$HOME") {
        return filepath.Join(user.HomeDir, path[5:])
    }
    return path
}
