{{- $render := . -}}
**App configuration on {{ .Config.GetName }}**

**Slashing params**
Blocks window: {{ .Config.BlocksWindow }}
Validator needs to sign {{ .FormatMinSignedPerWindow }}%, or {{ .Config.GetBlocksSignCount }} blocks in this window.
Average block time: {{ .FormatAvgBlockTime }} seconds
Approximate time to go to jail when missing all blocks: {{ .FormatTimeToJail }}

**App config**
Storing blocks amount: {{ .Config.StoreBlocks }}
Missed blocks thresholds:
{{ range .Config.MissedBlocksGroups -}}
{{ .EmojiEnd }} {{ .Start }} - {{ .End }} ({{ $render.FormatGroupPercent . }})
{{ end }}
