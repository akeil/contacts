# Edit Contact

Nick         : {{ .NickName | join }}
Prefix       : {{ .Name.HonorificNames | join }}
First Name   : {{ .Name.GivenName | join }}
Last Name    : {{ .Name.FamilyName | join }}

Categories   : {{ .Categories | join }}

Title        : {{ .Title }}
Role         : {{ .Role }}
Organization : {{ .Org }}

# Mail Adresses ---------------------------------------------------------------
# Format is     TYPE[, TYPE]: ADDRESS
# Types are     work, home
{{ range .Email }}
{{ .Type | join }}: {{ .Value }}{{ end }}
home:

# Phone Numbers ---------------------------------------------------------------
# Format is     TYPE[, TYPE]: NUMBER
# Types are:    text, voice, fax, cell, video, pager, textphone
{{ range .Telephones }}
{{ .Type | join }}: {{ .Value }}{{ end }}
voice:

# URLs ------------------------------------------------------------------------
# Format is     TYPE[, TYPE]: URL
# Types are     work, home
{{ range .Url }}
{{ .Type | join }}: {{ .Value }}{{ end }}
home:

# Postal Addresses ------------------------------------------------------------
# Format is     TYPE: ?; ?; STREET; CITY; REGION; POSTAL_CODE; COUNTRY
# Types are     work, home
{{ range .Addresses }}
{{ .Type | join }}: ; ; {{ .Street }}; {{ .Locality }}; {{ .Region }}; {{ .PostalCode }}; {{ .CountryName}}{{ end }}
#home: ; ; ; ; ; ;
