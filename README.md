# missed-blocks-checker

![Latest release](https://img.shields.io/github/v/release/QuokkaStake/missed-blocks-checker)
[![Actions Status](https://github.com/QuokkaStake/missed-blocks-checker/workflows/test/badge.svg)](https://github.com/QuokkaStake/missed-blocks-checker/actions)

missed-blocks-checker is a tool that tracks all validators' missed blocks and reports them
to reporters of your choice, along with other validators' actions (such as, tombstone, jail, unjail etc.).

## Why is it cool?
- Can work with multiple chain
- Can subscribe to multiple nodes on each chain, so if one goes down it'll continue to work
- Uses SQLite as a database to store all data on one place
- Is easily extendable to support other reporters or other chains
- Should work with all cosmos-sdk based chains out of the box
- Only RPC node is required as a data source, so no need for LCD nodes

## How can I set it up?

Download the latest release from [the releases page](https://github.com/QuokkaStake/missed-blocks-checker/releases/). After that, you should unzip it, and you are ready to go:

```sh
wget <the link from the releases page>
tar <downloaded file>
./missed-blocks-checker --config <path to config>
```

Alternatively, install `golang` (>1.18), clone the repo and build it:
```
git clone https://github.com/QuokkaStake/missed-blocks-checker
cd missed-blocks-checker
# This will generate a `missed-blocks-checker` binary file in the repository folder
make build
# This will generate a `missed-blocks-checker` binary file in $GOPATH/bin
```

To run it detached, first we have to copy the file to the system apps folder:

```sh
sudo cp ./missed-blocks-checker /usr/bin
```

Then we need to create a systemd service for our app:

```sh
sudo nano /etc/systemd/system/missed-blocks-checker.service
```

You can use this template (change the user to whatever user you want this to be executed from. It's advised to create a separate user for that instead of running it from root):

```
[Unit]
Description=Missed Blocks Checker
After=network-online.target

[Service]
User=<username>
TimeoutStartSec=0
CPUWeight=95
IOWeight=95
ExecStart=missed-blocks-checker --config <config path>
Restart=always
RestartSec=2
LimitNOFILE=800000
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

Then we'll add this service to autostart and run it:

```sh
sudo systemctl daemon-reload # reload config to reflect changed
sudo systemctl enable missed-blocks-checker # put service to autostart
sudo systemctl start missed-blocks-checker # start the service
sudo systemctl status missed-blocks-checker # validate it's running
```

If you need to, you can also see the logs of the process:

```sh
sudo journalctl -u missed-blocks-checker -f --output cat
```

## How does it work?

It subscribes to the new blocks on each chain specified in config, then on each new block
it queries the validators list and the signing infos, then reports to the configured reporters
if there are new events (like validator skipping blocks or getting jailed)

## How can I configure it?

All configuration is done via `.toml` config file, which is mandatory. Run the app with `--config <path/to/config.toml>` to specify config. Check out `config.example.toml` to see the params that can be set.

## Notifiers

Currently, this program supports the following notifications channels:
1) Telegram

Go to @BotFather in Telegram and create a bot. After that, there are three options:
- you want to send messages to a user. This user should write a message to @getmyid_bot, then copy the `Your user ID` number. Also keep in mind that the bot won't be able to send messages unless you contact it first, so write a message to a bot before proceeding.
- you want to send messages to a channel. Write something to a channel, then forward it to @getmyid_bot and copy the `Forwarded from chat` number. Then add the bot as an admin.
- you want to send message to a chat. Add @raw_data_bot to this chat, write something, then copy a channel_id from bot response (starts with a minus), then you can remove @raw_data_bot from the channel.

To have fancy commands auto-suggestion, go to @BotFather again, select your bot -> Edit bot -> Edit description and paste the following:
```
start - Displays bot info
help - Displays bot info
subscribe - Subscribe to a validator's updates
unsubscribe - Unsubscribe from a validator's updates
status - See missing blocks of validators you are subscribed to
validators - See missing blocks of all validators
missing - See validators who are missing blocks
notifiers - See notifiers for each validator
params - See chain and config params
config - See chain and config params
```

Then add a Telegram config to your config file (see `config.example.toml` for reference).

## How can I contribute?

Bug reports and feature requests are always welcome! If you want to contribute, feel free to open issues or PRs.
