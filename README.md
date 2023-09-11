# ntfy-bridge

Bridge for various implementations to publish to ntfy.

## Installation

Using go:

```sh
go install forge.babariviere.com/babariviere/ntfy-bridge@latest
```

Or using docker:

```sh
docker pull forge.babariviere.com/babariviere/ntfy-bridge:latest
```

Binaries are also avaiable in the [release section](https://forge.babariviere.com/babariviere/ntfy-bridge/releases).

## Usage

First, you need to create a configuration file. A sample one is provided [here](./config.example.scfg).

For now, we have these handler types:
- `flux`: handle notification from [Flux](https://fluxcd.io)
- `discord_embed`: handle preformated notification from discord embeds (see [embed object](https://discord.com/developers/docs/resources/channel#embed-object))
- `alertmanager`: handle notification from alertmanager using [webhook_config](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config)

Once you have created your config file, you can either put it in these directories:
- `/etc/ntfy-bridge/config.scfg`
- `$HOME/.ntfy-bridge/config.scfg`
- `$HOME/.config/ntfy-bridge/config.scfg`
- `config.scfg` (current directory)

Then, you can simply run the binary with either the native binary:

```sh
./ntfy-bridge
```

Or via docker:

```sh
docker run -v config.scfg:/etc/ntfy-bridge/config.scfg -p 8080 forge.babariviere.com/babariviere/ntfy-bridge:latest
```

Sample config for kubernetes can be found in [./k8s/](./k8s/) directory.
