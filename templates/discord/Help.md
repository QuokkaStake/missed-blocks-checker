[missed-blocks-checker](https://github.com/QuokkaStake/missed-blocks-checker">missed-blocks-checker)

This bot can monitor missing blocks for validators on multiple Cosmos chains,
subscribing to the notifications on multiple validators, and many more.

Created by [🐹 Quokka Stake](https://quokkastake.io) with ❤️.

The bot can understand the following commands:
- </help:{{ .help.Info.ID }}> - display this message
- /subscribe [validator address] - subscribe to validator's notifications
- /unsubscribe [validator address] - unsubscribe from validator's notifications
- /status - see the notification on validators you are subscribed to
- /missing - see the missed blocks counter of validators missing blocks
- /validators - see the missed blocks counter of all validators
- /config, or /params - see the app config and chain params
