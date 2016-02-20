--------------------[ {{ .FormattedName }} ]--------------------
{{- if .NickName }}
Nick         : {{ .NickName | join }}
{{- end }}
{{- if .Name.HonorificNames }}
Prefixes     : {{ .Name.HonorificNames | join }}{{ end }}
First Name   : {{ .Name.GivenName | join }}
Last Name    : {{ .Name.FamilyName | join }}
{{- if (len .Categories) gt 0 }}

Categories   : {{ .Categories | join }}
{{- end }}
{{- if .Title }}

Title        : {{ .Title }}
{{- end }}
{{- if .Role }}
Role         : {{ .Role }}
{{- end }}
{{- if .Org }}
Organization : {{ .Org }}
{{- end }}
{{- if ( len .Email ) gt 0 }}

Mail Adresses:
{{- range .Email }}
- [{{ .Type | join }}] {{ .Value }}
{{- end }}
{{- end }}
{{- if ( len .Telephones ) gt 0 }}

Phone Numbers:
{{- range .Telephones }}
- [{{ .Type | join }}] {{ .Value }}
{{- end }}
{{- end }}
{{- if (len .Url) gt 0 }}

URLs:
{{- range .Url }}
- [{{ .Type | join }}] {{ .Value }}
{{- end }}
{{- end }}
{{- if ( len .Addresses ) gt 0 }}

Adresses:
{{- range .Addresses }}
- [{{ .Type | join }}]
  {{.Street  }}
  {{ .PostalCode }} {{ .Locality }}
  {{ .CountryName }}
{{- end }}
{{- end }}
{{- if .Note }}

Notes:
{{ .Note}}
{{- end }}
---------------------------------------------------
