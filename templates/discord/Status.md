{{- $render := . -}}
You are subscribed to the following validators' updates on {{ .ChainConfig.GetName }}:
{{- range .Entries }}
{{ if .Validator.Jailed -}}
**{{ SerializeLink .Link }}:** jailed
{{- else if not .Validator.Active -}}
**{{ SerializeLink .Link }}:** not in the active set
{{- else if .Error -}}
**{{ SerializeLink .Link }}:**: error getting validators missed blocks: {{ .Error }}
{{- else -}}
**{{ SerializeLink .Link }}** ({{ $render.FormatVotingPower . }}): {{ .SigningInfo.GetNotSigned }} missed blocks ({{ $render.FormatNotSignedPercent . }}%)
{{- end -}}
{{ end }}