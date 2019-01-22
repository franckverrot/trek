package main

const (
	jobsListFormat          = `{{range .Jobs}}* {{.Name}}{{println}}{{end}}`
	allocationsFormat       = `{{range .Allocations}}* {{.Name}}{{println}}{{end}}`
	allocationDetailsFormat = `{{range $index, $task := .Tasks}}({{$index}}) {{$task.Name}}{{println}}{{end}}`
	taskGroupsListFormat    = `{{range .TaskGroups}}* {{.Name}}{{println}}{{end}}`
	taskDetailsFormat       = `{{- "" -}}
* Name: {{ .Task.Name }}
* Node Name: {{ .Node.Name }}
* Node IP: {{ .Node.IP }}
* Driver: {{ .Task.Driver }}{{println}}
	{{- range $configKey, $configValue := .Task.Config -}}
		{{"  * "}}{{$configKey}}: {{$configValue}}{{println}}
	{{- end -}}
{{- if .Environment -}}
* Env:{{println}}
	{{- range $key, $value := .Environment -}}
		{{"  * "}}{{$key}}: {{$value.Value}}{{println}}
	{{- end -}}
{{- end -}}
{{- if .Network.ReservedPorts -}}
* Reserved Ports:{{println}}
	{{- range .Network.ReservedPorts -}}
		{{"  * "}}{{.Number}} ({{.Name}}){{println}}
	{{- end -}}
{{- end -}}
{{- if .Network.DynamicPorts -}}
* Dynamic Ports:{{println}}
	{{- range .Network.DynamicPorts -}}
		{{"  * "}}{{.Number}} ({{.Name}}){{println}}
	{{- end}}
{{- end -}}
{{- "" -}}`
)
