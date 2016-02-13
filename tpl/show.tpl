--------------------[ Contact ]--------------------{{if .NickName }}
Nick      : {{ .NickName | join }}{{end}}
First Name: {{ .Name.GivenName | join }}
Last Name : {{ .Name.FamilyName | join }}
{{ if ( len .Email ) gt 0 }}
Mail Adresses:
{{ range .Email }} - [{{ .Type | join }}] {{ .Value }}
{{ end }}{{ end }}
{{ if ( len .Telephones ) gt 0 }}Phone Numbers:
{{ range .Telephones }} - [{{ .Type | join }}] {{ .Value }}
{{ end }}{{ end }}
{{ if ( len .Addresses ) gt 0 }}Adresses:
{{ range .Addresses }}  [{{ .Type | join }}]
  {{.Street  }}
  {{ .PostalCode }} {{ .Locality }}
  {{ .CountryName }}{{end}}
{{ end }}
---------------------------------------------------