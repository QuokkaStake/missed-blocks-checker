{{- if . }}
**Showing the last {{ len .Events }} entries for {{ SerializeLink .ValidatorLink }}:**
{{- range .Events }}
{{ .FormatTime }}: {{ .RenderedEvent }}
{{- end }}
{{- else }}
No events for this address since the app launch.
{{- end }}