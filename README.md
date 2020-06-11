# Minecraft Server Properties API

Find useful information about Minecraft's Java Edition server configuration file, server.properties, with this simple, RESTful API!

## Usage

Make a simple GET request and decode the JSON response!

To get the whole documentation: `GET http://api.mcbonanza.games/serverproperties/v1/`<br>
To get a single key: `GET http://api.mcbonanza.games/serverproperties/v1/{key}`, where `{key}` is a valid server.properties key.

### Response examples

- **Whole documentation** (on `GET http://api.mcbonanza.games/serverproperties/v1/`)

```json
{
  "properties": [
    {
      "name": "allow-flight",
      "type": "boolean",
      "defaultValue": "false",
      "values": {
        "min": 0,
        "max": 1,
        "possible": ["false", "true"]
      },
      "description": "Allows users to ...", // complete description
      "upcoming": false,
      "upcomingVersion": ""
    }
    // rest of keys
  ]
}
```

- **Single key** (on `GET http://api.mcbonanza.games/serverproperties/v1/difficulty`)

```json
{
  "name": "difficulty",
  "type": "string",
  "defaultValue": "easy",
  "values": {
    "min": -2147483648,
    "max": -2147483648,
    "possible": ["peaceful", "easy", "normal", "hard"]
  },
  "description": "Defines the difficulty ...", // complete description
  "upcoming": false,
  "upcomingVersion": ""
}
```

- **Error** (on `GET http://api.mcbonanza.games/serverproperties/v1/sarmale`). The errors are appended to the array in the order they occur.

```json
{
  "errors": [
    {
      "error": "404 Not Found, key \"sarmale\" doesn't exist",
      "retry": false
    }
  ]
}
```

## How it works

The API scrapes the official Minecraft Wiki page (specifically the [server.properties page](https://minecraft.gamepedia.com/Server.properties)), also evaluating
the mentioned values and limits a property can be assigned. The math expressions are evaluated using [math.js web service](https://api.mathjs.org/). All the
properties are then stored in a list, which is accessed on each client request.

## Contributions

Fork this project, create a branch from develop, do your work, and open a pull request in the origin's develop branch.
Don't forget to get the module's dependencies! Use goimports to format the code and LF line endings.
