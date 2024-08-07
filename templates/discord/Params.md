{{- $render := . -}}
**App configuration on {{ .Config.GetName }}**

**Slashing params**
Blocks window: {{ .Config.BlocksWindow }}
Validator needs to sign {{ .FormatMinSignedPerWindow }}%, or {{ .Config.GetBlocksMissCount }} blocks in this window.
Average block time: {{ .FormatAvgBlockTime }} seconds
Approximate time to go to jail when missing all blocks: {{ .FormatTimeToJail }}

**Chain info**
{{ if .Config.IsConsumer.Bool -}}
The chain is an ICS consumer chain.
{{- else -}}
The chain is a sovereign chain.
{{- end }}

**App config**
Interval between sending/generating reports: {{ .FormatSnapshotInterval }}
Missed blocks thresholds:
{{ range .Config.MissedBlocksGroups -}}
{{ .EmojiEnd }} {{ .Start }} - {{ .End }} ({{ $render.FormatGroupPercent . }})
{{ end }}
