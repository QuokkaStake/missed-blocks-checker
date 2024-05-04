{{- if not .Validators }}
There are no active validators on {{ .Config.GetName }}!
{{- else }}
**Validators' status on {{ .Config.GetName }}:**
{{- end }}
{{ range .Validators -}}
**{{ SerializeLink .Link }}**: {{ .NotSigned }} missed blocks ({{ .FormatMissed }}%)
{{ end }}
