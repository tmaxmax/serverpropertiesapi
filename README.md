| :exclamation: The v1 API has been deprecated, endpoints are no longer available! Use v2 instead! |
| ------------------------------------------------------------------------------------------------ |


# Minecraft Server Properties API

Find useful information about Minecraft's Java Edition server configuration file, server.properties, with this simple, RESTful API!

## Usage

Make a simple GET request and decode the JSON response!

To get the whole documentation: `GET http://api.mcbonanza.games//v1serverproperties`  
To get a single key: `GET http://api.mcbonanza.games/serverproperties/v1/{key}`, where `{key}` is a valid server.properties key.  
To get metadata (such as the default limit value and the possible value types): `GET http://api.mcbonanza.games/v1/serverproperties/meta/`

### Response examples

- **Whole documentation** (on `GET http://api.mcbonanza.games/v1/serverproperties`)

```json
{
  "options": {
    "contains": [],
    "type": [],
    "upcoming": ""
  },
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
      "description": "Allows users to ...",
      "upcoming": false,
      "upcomingVersion": ""
    }
    ...
  ]
}
```

You can filter your requests with the provided options!

- `contains` option only returns the properties that contain the substring the key is attributed
- `type` option only returns the properties of the requested type
- `upcoming` option only returns properties that are going to be implemented in future versions, if `"true"`, or only properties currently available, if `"false"`

_Example request_: `GET http://api.mcbonanza.games/v1/serverproperties?upcoming=true&type=integer`

```json
{
  "options": {
    "contains": "",
    "type": "integer",
    "upcoming": "true"
  },
  "properties": [
    {
      "name": "entity-broadcast-range-percentage",
      "type": "integer",
      "defaultValue": "100",
      "values": {
        "min": 0,
        "max": 500,
        "possible": []
      },
      "description": "Controls how close ...",
      "upcoming": true,
      "upcomingVersion": "JE 1.16"
    }
  ]
}
```

If there is no property that matches your options, an empty array will be returned.

- **Single key** (on `GET http://api.mcbonanza.games/v1/serverproperties/difficulty`)

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
  "description": "Defines the difficulty ...",
  "upcoming": false,
  "upcomingVersion": ""
}
```

- **Error** (on `GET http://api.mcbonanza.games/v1/serverproperties/sarmale`). The errors are appended to the array in the order they occur.

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
the mentioned values and limits a property can be assigned. The math expressions are evaluated using [math.js web service](https://api.mathjs.org/). Then the data is sent to the user, caching it if possible.

## Contributions

Fork this project, create a branch from develop, do your work, and open a pull request in the origin's develop branch.
Don't forget to get the module's dependencies! Use goimports to format the code and LF line endings.
