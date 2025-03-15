# HLL Discord Server Watcher

This tool will poll changes from a control panel of a supported game service provider for the game Hell Let Loose.
It will try to fetch the current server name and server password and will publish both information in a configurable Discord channel.

The main idea is to ensure that members of a discord server (most likely HLL clan) always have access to the current server password should this change often.
It reduces the manual effort server maintainers, event planers or others need to make when they change the server password (= remembering to post the password in discord).

# Installation

Easiest way is to use the provided `docker-compose.example.yml` file and host the tool as a docker container:

```shell
git clone git@github.com:FlorianSW/hll-discord-server-watcher.git
cd hll-discord-server-watcher
cp docker-compose.example.yml docker-compose.yml
touch config.json
# now is the time to fill the config.json with the necessary configuration properties, see following sections
docker compose up -d
```

# Configuration

All configuration is done in a file called `config.json`.

Create a new file with that name in the root directory of the project and copy & paste this content there:
```json
{
  "discord": {
    "token": "your_bot_token",
    "guild": "your_guild_id",
    "channel_id": "your_channel_id"
  },
  "control_panel_base_url": "qp.qonzer.com",
  "servers": [
    {
      "name": "custom_name",
      "color": 2123412,
      "service_id": "service_id_in_gsp_panel",
      "credentials": {
        "username": "username_in_gsp_panel",
        "password": "password_in_gsp_panel"
      }
    }
  ]
}
```

## Discord related settings

First, create a new discord application in the [discord developer portal](https://discord.com/developers/applications).
Invite this discord bot to your discord server by:
- Navigate to _OAuth2_
- In the _OAuth2 URL Generator_ section, select `bot`
- In the newly opened _Bot Permissions_ section, select `Send Messages`
- Copy the `Generated URL` url from the bottom section and open that in a new browser tab
- Follow the invite screen and add the bot to your server

In your discord server, make sure that the discord bot has access (_View Channel_ and _Send Messages_) to the password channel were it should publish the messages to.

Afterward, generate a new bot token in the discord developer portal:
- Navigate to _Bot_
- Under _Token_ select _Reset token_
- Copy the token and replace it in your config.json

In your discord client, enable the [developer mode](https://discord.com/developers/docs/activities/building-an-activity#step-0-enable-developer-mode).
Then right-click on your discord server and hit _Copy Server ID_ and paste it into your `config.json` at the `guildId` key.
Do the same for the channel where the servers should be posted to.

## Control Panel of GSP settings

Put the base url of your GSPs control panel into the `control_panel_base_url`, e.g. `qp.qonzer.com` for Qonzers control panel.

Navigate to the control panel and create a new sub-user with access to _Configuration files_ on all servers you want to get information from.
Then, add the username, password and service ID (which can be found in the URL when you navigate to a server in the control panel) to your `config.json`.
You can add more than one server by duplicating the server object and separate them with commas, e.g.:
```json
{
  // ...
  "servers": [
    {
      "name": "custom_name",
      "color": 2123412,
      "service_id": "service_id_in_gsp_panel",
      "credentials": {
        "username": "username_in_gsp_panel",
        "password": "password_in_gsp_panel"
      }
    }, {
      "name": "custom_name_2",
      "color": 2123412,
      "service_id": "service_id_in_gsp_panel",
      "credentials": {
        "username": "username_in_gsp_panel",
        "password": "password_in_gsp_panel"
      }
    }
  ]
}
```

The value in `color` is a color code of Discord.
Use the int value of [these color codes](https://gist.github.com/thomasbnt/b6f455e2c7d743b796917fa3c205f812).
