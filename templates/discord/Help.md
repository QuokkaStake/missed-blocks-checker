[missed-blocks-checker](<https://github.com/QuokkaStake/missed-blocks-checker>) v{{ .Version }}

This bot can monitor missing blocks for validators on multiple Cosmos chains,
subscribing to the notifications on multiple validators, and many more.

Created by [🐹 Quokka Stake](<https://quokkastake.io>) with ❤️.

The bot can understand the following commands:
- </help:{{ .Commands.help.Info.ID }}> - display this message
- </subscribe:{{ .Commands.subscribe.Info.ID }}> [validator address] - subscribe to validator's notifications
- </unsubscribe:{{ .Commands.unsubscribe.Info.ID }}> [validator address] - unsubscribe from validator's notifications
- </status:{{ .Commands.status.Info.ID }}> - see the notification on validators you are subscribed to
- </missing:{{ .Commands.missing.Info.ID }}> - see the missed blocks counter of validators missing blocks
- </validators:{{ .Commands.validators.Info.ID }}> - see the missed blocks counter of all validators
- </params:{{ .Commands.params.Info.ID }}> - see the app config and chain params
- </notifiers:{{ .Commands.notifiers.Info.ID }}> - see notifiers for each validator
- </jails:{{ .Commands.jails.Info.ID }}> - see latest jails and tombstones events
- </events:{{ .Commands.events.Info.ID }}> [validator address] - see latest events for a validator
- </jailscount:{{ .Commands.jailscount.Info.ID }}> - see jails count for each validator since the app was started
