# Installation

## Natively

First thing, we need to install some packages

```Bash
sudo apt install build-essential golang-go git yt-dlp ffmpeg libopus-dev -y
```

After installation is done, we can clone the repo with

```Bash
git clone https://github.com/TheTipo01/YADMB
```

Enter the directory, and build the bot

```Bash
cd YADMB
go build
```

We only need to install DCA

```Bash
go get -u github.com/bwmarrin/dca/cmd/dca
```

Final things:
- add `dca` to your path, you can do that by creating a symlink of that executable to your `/usr/bin` directory (`ln -s /home/thetipo01/go/bin/dca /usr/bin/dca`)
- modify the `example_config.yml`, adding all required tokens and renaming it to `config.yml`
- for info about creating and adding the bot, see the following [page](hosting.md)
## Docker

- Clone the repo
- Modify the `example_config.yml`, by adding your discord bot token (
  see [here](hosting.md) if you don't know how
  to do it)
- Rename it in `config.yml` and move it in the `data` directory
- Run `docker-compose up -d`
- Enjoy your YADMB instance!

Note: the docker image is available
on [Docker hub](https://hub.docker.com/r/thetipo01/yadmb), [Quay.io](https://quay.io/repository/thetipo01/yadmb)
and [Github packages](https://github.com/TheTipo01/YADMB/pkgs/container/yadmb).
