Nick         : {{ .NickName | join }}
Prefix       : {{ .Name.HonorificNames | join }}
First Name   : {{ .Name.GivenName | join }}
Last Name    : {{ .Name.FamilyName | join }}
Title        : {{ .Title }}
Role         : {{ .Role }}
Organization : {{ .Org }}

# Mail Adresses ---------------------------------------------------------------
# Format is TYPE[, TYPE]: ADDRESS
# Types are WORK, HOME
{{ range .Email }}
{{ .Type | join }}: {{ .Value }}{{ end }}
home:

# Phone Numbers ---------------------------------------------------------------
# Format is TYPE[, TYPE]: NUMBER
# Types are WORK, HOME, CELL
{{ range .Telephones }}
{{ .Type | join }}: {{ .Value }}{{ end }}
home:

# URLs ------------------------------------------------------------------------
# Format is TYPE[, TYPE]: URL
# Types are HOME, WORK
{{ range .Url }}
{{ .Type | join }}: {{ .Value }}{{ end }}
home:

# Postal Addresses ------------------------------------------------------------
# Format is "TYPE: ?; ?; STREET; CITY; REGION; POSTAL_CODE; COUNTRY"
# Types are WORK, HOME
{{ range .Addresses }}
{{ .Type | join }}: ; ; {{ .Street }}; {{ .Locality }}; {{ .Region }}; {{ .PostalCode }}; {{ .CountryName}}{{ end }}
