{{- if not .Validators }}
There are no missing validators on {{ .Config.GetName }}!
{{- else }}
**Validators missing blocks on {{ .Config.GetName }}:**
{{- end }}
{{ range .Validators -}}
**{{ SerializeLink .Link }}**: {{ .NotSigned }} missed blocks ({{ .FormatMissed }}%)
{{ end }}
