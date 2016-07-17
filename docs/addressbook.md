# Address Book
A [khard](https://github.com/scheibler/khard/) clone

## Features


### UI
command line only,
maybe an interactive shell (what for?)

To edit a contact, use a template file that is filled with the contact details
and opened in an editor.
When the editor is closed (and the file is changed), parse the files and
transfer contact data to the VCF file.


### Basics
* add a contact
* update a contact
* remove a contact
* list contacts
* search contacts by name
* search contacts by tag


### Extra
* merge contacts(?)
* move contacts between address books
* full-text search (with whoosh?)
* normalize *vCard* by removing unwanted properties
* export
    * list (of search results)
    * single contact
    * as VCF
    * as JSON?
    * as CSV
    * as shell-vars
* support for `KIND:group` vCards
* support for `CATEGORIES` (tags
* plugin for exporting to custom format
(e.g. contacts.txt for sup)
* DBus interface (maybe a standard exists?)


## Commands


### add
adds a contact
```
-
```

### ls
lists contacts

    $ contacts ls

Should accept a filter expression as a positional argument:

    $ contacts ls doe

To find every contact containing "doe" (case insensitive, "jo" finds "john")
search fields are:
- name
- first name, last name
- nickname
- mail addresses

Output is a tabulated list with one row per contact.
The list contains some sane defaults
- full name
- primary mail
- primary phone

One can customize the columns to show
either on the command line or in a config file.

Sort order is by name ("human sort", A comes before b)
order is also customizable through CLI or config.

Should also be possible to use `ls` to generate output for other
commands. Thus, we need different formatters:
- CSV
- JSON
- Shell Args
- VCF

`ls` might also support more complex filter expressions
where one can query specific fields or groups of contacts.

Use e.g. to generate lookup lists for your mail client.

### show
Show details for one contact

    $ contacts show john

Accepts a filter/query as a positional arg.
if multiple matches are found, and if we are interactive,
display a short list with matches, user selects one.

### rm
delete a contact

    $ contacts rm john

Again with matchin by name.
when multiple contacts are matched, ask which one(s) to delete.
Allow to confirm any number of contacts to delete.

As an option, do not delete but move to "Trash".

### merge
Merge data from two or more contacts.

### diff
show difference between two contacts.

### dedup
Find duplicate contacts and offer to merge or delete them.


## Persistence
Store contacts as `vcf` files, one file per contact.
A directory constitutes an *Addressbook* (one can have multiple address books).

Fully parse and write vcf
use [vobject](https://pypi.python.org/pypi/vobject) for this

But also: drop a vcf file into the directory and be done.

Would be nice if VCF files had pretty names derived
from the contacts name. E.g.:

    john-doe.vcf

(And *not* use a generated GUID)


## Classes


### Contact
* first name
* last name
* nick name
* mail addresses
* phone numbers
* postal addresses
* birthday

Include calculated properties for the *preferred*
phone number, mail address, etc.
The *preferred* properties have the `PREF` parameter
(add own method to always produce a preferred value).


### Addressbook
* directory
* list of contacts ("collection")


### Template
The template that is used for editing a contact.
* read - load from file
* render - write contact properties into the template
* parse - parse edits from the template


### Renderer
Use to render console output.
Somewhat related to *Template* as we want to format
the same data.


### Edit Controller
Maybe.
Code to handle edits.


## Configuration
* editor to use (default ENV `$EDITOR` and `$VISUAL`)
* addressbook:
    * path where the VCF's are stored
    * name (normally, the name of the directory)

## vCard Fields
[vCard format spec](https://tools.ietf.org/html/rfc6350)
[vCard MIME directory profile](https://www.ietf.org/rfc/rfc2426.txt)
a VCF file has the following fields:

```
Field       | C. | Description
------------|----|------------------------------------
FN          | 1* | full name, display name?)
N           | *1 | Name
NICKNAME    | *  | Nickname
PHOTO       | *  | a URI or base64 encoded
BDAY        | *1 | the birthday
ANNIVERSARY | *1 | date of marriage
GENDER      | *1 | (M)ale, (F)emale, (O)ther, ...
ADR         | *  | Postal addresses
TEL         | *  | phone numbers; TYPE, PREF
EMAIL       | *  | mail addresses; TYPE, PREF
LANG        | *  | languages; TYPE, PREF
TZ          | *  | timezone
TITLE       | *  | job title
ROLE        | *  | organizational role
LOGO        | *  | company(?) logo, URI
ORG         | *  | organization and org. unit
CATEGORIES  | *  | list of tags
NOTE        | *  | free-form text note
```

Cardinality:
1  = one instance required
*1 = one instance, optional
1* = one or more required
\* = any number


### N (Name)
The name components, separated by ";"
within a component, separate by ","

1. family names (surname, last name)
2. given names (first name)
3. additional names (?)
4. prefixes
5. suffixes

Example:
```
N:Doe;John
N:Doe;John;Paul;Dr.,Prof.;Jr.
```



### TEL (Phone)
RFC6350 says, this *should* by a URI with the `tel` scheme.
But it can also be a free-form text.

Maybe try to parse that into a tel-URI
[RFC 2966](https://tools.ietf.org/html/rfc3966) describes the tel URI:

    tel:+49-221-123-456

Phone numbers can have the `PREF` parameter
to indicate the preferred number.

Phone numbers can also have one or more `TYPE`
parameters from this list:

* text (=supports SMS)
* voice
* fax
* cell
* video
* pager
* textphone


### EMAIL (Mail)
Single text value, but should be a valid mail address.

Can have a `PREF` parameter that marks the preferred
address.

Mails can have a `TYPE` parameter like

* work
* more?


### BDAY (Birthday)
The Birthday, formatted as `yyyymmdd`, e.g. 19700131.


## Editor Template
Editing a contact is done by loading contact details into
a template text file. One can then change the properties
of the contact and save the file. Upon save, application
parses the file and updates the `VCF` for that contact.

The template file can be user-supplied.

The same mechanism can be used to create a new contact.

The template looks like this:

Lines starting with a `#` are **comments**:
```
 # This is a comment
```

Simple properties (those that have a 1:1 relationship to
a contact) are expressed with `key: ?` pairs:
```
Name: ?
```

Some rules:

* The property name must occur at the start of the line.
* There must be now other characters after the ":"
  (which means we could ditch the ?)
* The question mark will be replaced by the value of
  the *Name* property.
* Property names are case **insensitive**.

* Property names may also **include spaces**. So

    First Name

is equivalent to

    FirstName

* Maybe, also ignore dash ("-") and underscore ("_"),
if they are not part of vCard field names.

* Values are stripped of spaces, so one can align them:
```
First Name :
Last Name  :
Nick Name  :
```

If the template has fields **missing**, the values for
these fields are not rendered and cannot be edited.
However, they will not be lost during edit.


### Help Text
Since the values will be parsed and validated,
some benefit from a help text.
A template may either contain help texts verbatim
or use placeholders for help:
```
help.address
Address:
```
Help will be rendered as a comment
(one ore more lines).


### Aliases
Some of the vCard fields have not-so-pretty names.
We alias them:

N - Name
FN - Full Name
TEL - Phone


### Special Cases
The vCard specifies two Name fields:
`N` and `FN`.

N consists of components like "Last Name", "First Name".
Split these into multiple fields and combine them
into the N components.

Calculate the FN value from the first N field.

### Cardinality
Many fields are allowed multiple times.
In this case, the template engine:

* may repeat the line with that property when rendering
directly below the first occurrence
* parse every occurrence of that property


### Multiline Values
Users can enter a value spanning multiple lines.
The template engine should wrap long lines
(use a line-wrap config option).

A wrapped line looks like this:
```
Note: This is a note which
spans across multiple lines
```


But also like this (the leading whitespace is stripped):
```
Note: This is a note which
      spans across multiple lines
```


### Types
For parsing an rendering, the template engine must
associate a value *type* to each property and apply
the correct formatter/parser.
This allows e.g. to present dates in a user-preferred
format instead of using them 1:1 from the VCF

Some values consist of several components,
separated by a semicolon (";"). In cases where it is
*common* to use multiple components, the ";" may be
rendered into the template.
When parsing, missing components are appended with
empty values.


### Parameters
Some fields can have parameters. A parameterized field
looks like this:

    EMAIL;TYPE=work:jdoe@example.tld

Parameters come after the field name and before the value.
One can have multiple parameters on a field.

The Template form of a parameter looks like this
```
Mail (type=work): jdoe@example.tld
```
Alternatives:
```
Mail (work): jdoe@example.tld
Mail       : (work) jdoe@example.tld
Mail       : jdoe@example.tld (work)
Mail, work : jdoe@example.tld
```

One could include special handling for the `PREF`
parameter by applying the PREF to and from the
sequence.


### Example

Here is a sample template:
```
First Name:
Last Name :
Nickname  :

\# phone numbers
Phone:

\# mail addresses
Mail:

\# postal addresses
Address:
```

The rendered template may look like this:
```
First Name: John
Last Name : Doe
Nickname  : Johnny

\# phone numbers
Phone: +49 123 465 7890
Phone: +49 100 200 3000

\# mail addresses
Mail: john.doe@example.tld

\# postal addresses
Address: Main Street 1; 12345; Some City; ; Some Country
```
