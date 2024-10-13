{{- if not .Entries }}
Nobody is subscribed to any notifications on {{ .Config.GetName }}!
{{- else }}
**Validators' notifiers on {{ .Config.GetName }}:**
{{- end }}
{{ range .Entries -}}
- **{{ SerializeLink .Link }}**: {{ SerializeNotifiersNoLinks .Notifiers }}
{{ end }}
