package contacts

import (
    "encoding/json"
    "io"
    "log"
    "os"
)


type Configuration struct {
    Addressbook string
    Editor string
}

func (s Configuration) merge(other Configuration) {
    if other.Addressbook != "" {
        s.Addressbook = other.Addressbook
    }
    if other.Editor != "" {
        s.Editor = other.Editor
    }
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
        //config.merge(cfg)
    }
    return config
}
