# Addressbook
A Command line addressbook that works with a collection of
[vCard](https://tools.ietf.org/html/rfc6350) files in a single directory.


## Usage
Use the following commands:

To `add` a new contact:

```
$ card add -f John -l Doe
```
Will add a new contact to the addressbook and open it in the editor.

To `edit` an existing contact:
```
$ card edit john
```
The *edit* command takes a single search term. If that term matches exactly
one contact, that contact is opened in the editor.
IF multiple matches are found, one is chosen.

To `del`(ete) a contact:
```
$ card del john
```
As with *edit*, a single search term is used to select the contact to be
deleted.
Only one contact will be deleted at a time.

`ls` will produce a list of all contacts, optionally filtered by a search term.
```
$ card ls
```

To `show` details for a single contact:
```
$ card show john
```


## Configuration
Configuration is kept in JSON format at `~/.config/contacts.config.json`.
The configuration file looks like this:

``` json
{
    "Addressbook": "~/contacts",
    "Editor": "/usr/bin/nano"
}
```

- **Addressbook**: The path to a directory with vCards.
  This is where contacts are stored.
- **Editor**: An executable that is used to edit contacts.
  This should be a text editor.

## Similar Tools
- [khard](https://github.com/scheibler/khard/) offers the same functionality,
  written in Python.
- [vdirsyncer](https://github.com/untitaker/vdirsyncer/) can be used to sync
  two or more adressbooks.
