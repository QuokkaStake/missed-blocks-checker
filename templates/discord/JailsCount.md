{{- if . }}
**Validators jails count observed by the app:**
{{- range . }}
- {{ SerializeLink .ValidatorLink }}: {{ .JailsCount }}
{{- end }}
{{- else }}
Nobody has been jailed since the app launch.
{{- end }}