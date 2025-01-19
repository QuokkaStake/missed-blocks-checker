{{- if . }}
**Showing the last {{ len . }} entries:**
{{- range . }}
{{ .FormatTime }}: {{ .RenderedEvent }}
{{- end }}
{{- else }}
Nobody has been jailed since the app launch.
{{- end }}