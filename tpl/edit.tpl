Nick         : {{ .NickName | join }}
First Name   : {{ .Name.GivenName | join }}
Last Name    : {{ .Name.FamilyName | join }}
Title        : {{ .Title }}
Role         : {{ .Role }}
Organization : {{ .Org }}
URL          : {{ .URL }}

# Mail Adresses ---------------------------------------------------------------
# Format is TYPE[, TYPE]: ADDRESS
# Types are WORK, HOME

{{ range .Email }}{{ .Type | join }}: {{ .Value }}
{{ end }}

# Phone Numbers ---------------------------------------------------------------
# Format is TYPE[, TYPE]: NUMBER
# Types are WORK, HOME, CELL

{{ range .Telephones }}{{ .Type | join }}: {{ .Value }}
{{ end }}

# Postal Addresses ------------------------------------------------------------
# Format is "TYPE: ?; ?; STREET; CITY; REGION; POSTAL_CODE; COUNTRY"
{{ range .Addresses }}
{{ .Type | join }}: ; ; {{ .Street }}; {{ .Locality }}; {{ .Region }}; {{ .PostalCode }}; {{ .CountryName}}{{ end }}
