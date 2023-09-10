{{- $render := . -}}
**App configuration on {{ .Config.GetName }}**

**Slashing params**
Blocks window: {{ .Config.BlocksWindow }}
Validator needs to sign {{ .FormatMinSignedPerWindow }}%, or {{ .Config.GetBlocksMissCount }} blocks in this window.
Average block time: {{ .FormatAvgBlockTime }} seconds
Approximate time to go to jail when missing all blocks: {{ .FormatTimeToJail }}

**App config**
Missed blocks thresholds:
{{ range .Config.MissedBlocksGroups -}}
{{ .EmojiEnd }} {{ .Start }} - {{ .End }} ({{ $render.FormatGroupPercent . }})
{{ end }}
