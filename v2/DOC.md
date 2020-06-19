# Server Properties API documentation

This is MC Bonanza's _server.properties_ reference REST API documentation. This API can be used to retrieve information from the [official Minecraft Wiki](https://minecraft.gamepedia.com/) about the server.properties keys and values. In this documentation, the term _official documentation_ will refer to the [Minecraft Wiki server.properties documentation](https://minecraft.gamepedia.com/Server.properties), and all its different languages it's written in.

**Root:** `https://api.mcbonanza.games/v2/serverproperties`

### <a id="constants"></a>Constants/Metadata

- <a id="meta-limitNotComputed"></a>**Limit not computed:** `-2147483648` (int32 minimum number)  
  This integral value is used to signify that an integer property has no limits documented in the official documentation. Metadata key: `limitNotComputed`.
- <a id="meta-minecraftTypeNames"></a>**Minecraft type-names**  
  This string array represents the possible property value types specified by the API in the requested language.

### <a id="queries"></a>Request queries

The API accepts different URL queries in the formats `key1=val1,val2&key2=val` or `key1=val1&key1=val2&key2=val`. If an endpoint accepts queries, they will be documented there.

## <a id="header"></a>Request header

This section documents the request header keys that affect the response.

- <a id="header-accept"></a>**Accept**  
  For the request to be accepted, make sure your client accepts both `application/json` and `application/problem+json`. See also the [unaccepted type error](#error-020)
- <a id="header-accept-encoding"></a>**Accept-Encoding**  
  The API can return gzipped data. If your client supports it, include `gzip` in this key's value.

## <a id="endpoints"></a>Endpoints

- [GET](#endpoint-allproperties)
- [GET `/{key}`](#endpoint-property)
- [GET `/meta/`](#endpoint-meta)

### <a id="endpoint-allproperties"></a>GET

This endpoint is used to get all properties from the official documentation.

_Example request:_ `GET https://api.mcbonanza.games/v2/serverproperties`  
_Response:_

```json
[]
```

This is the list of accepted filters:

- **contains** filters the response, so the properties' names contain the passed values. For example, `contains=allow,enable` will filter properties that don't have `allow` and `enable` as substrings, leaving only properties such as `allow-nether` or `enable-rcon`. You can also filter out properties that contain a substring bu putting a `!` before the value (for example `contains=!max` will return all properties that don't contain `max`)
- **type** filters the response, so the properties' value types are only the passed values. To get all available types, request to the [meta endpoint](#endpoint-meta). Values can also be negated with `!`. Don't mix negated with positive (such as `type=integer,!string`) - an [invalid type query error](#error-010) is returned. A few examples:

  - `type=integer` will return only integral properties
  - `type=integer,string` will return both integral and string properties
  - `type=!boolean` will return properties that are not boolean
  - `type=!string&type=!integer` will return properties that aren't integral or string type (so will return boolean properties only, the request should be simplified to `type=boolean`)

- **upcoming** filters the response based on its boolean value. If it's `true`, the response contains only properties that will be implemented in future versions and mentioned by the official documentation, else it contains only currently implemented properties. If its value isn't either `true` or `false`, [invalid upcoming query error](#error-011) is returned.

If no element matches the filters, the array will be empty.

The response can also be sorted using the **sort** query:

- `sort=name` will sort the response ascending, lexicographically by property name
- `sort=type` will sort the response ascending, lexicographically by type
- `sort=upcoming` will move the upcoming features to the end of the list

You can inverse the sorting order by putting a `-` in front of the desired sorting category (for example, `sort=-name` will sort the response descending, lexicographically by property name).

Any other queries or unspecified query values are ignored by the API.

### <a id="endpoint-property"></a>GET `/{key}`

This endpoint is used to get a single property named exactly like the `key` parameter's value from the official documentation.

_Example request:_ `GET https://api.mcbonanza.games/v2/serverproperties/allow-nether`
_Response:_

```json
{}
```

If the property doesn't exist, the API returns [error 000](#error-000).

### <a id="endpoint-meta"></a>GET `/meta/`

This endpoint is used to get the [constants](#constants) used by the API and other data, such as property value types.

_Example request:_ `GET https://api.mcbonanza.games/v2/serverproperties/meta/`
_Response:_

```json
{}
```

## <a id="errors"></a>Errors

The API returns the first error encountered in a request in the [RFC 7087 format](https://tools.ietf.org/html/rfc7807#section-3.1). The client must also accept `application/problem+json` MIME-type as a response.

### <a id="error-000"></a>Error code 000 - Property not found

This error is returned when requesting an inexistent property. HTTP response code 404 (Bad request) is also returned.

### <a id="error-010"></a>Error code 010 - Invalid type query

This error is returned when the type query filter contains mixed allowed and disallowed values. HTTP response code 400 (Bad request) is also returned.

### <a id="error-020"></a>Error code 020 - Unaccepted type

This error is returned when the client requests MIME-types that the API does not support. To solve this error, make sure your `Accept` header key includes `application/json` and `application/problem+json`, or `*/*`. The MIME-types are also returned in the error's description. HTTP response code 406 (Not acceptable) is also returned.

### <a id="error-021"></a>Error code 021 - Method not allowed

This error is returned when the client makes a request using a disallowed HTTP method. HTTP response code 405 (Method not allowed) is also returned.

### <a id="error-022"></a>Error code 022 - Endpoint invalid or missing

This error is returned when the client requests an invalid or missing endpoint. HTTP response code 404 (Not found) is also returned.

### <a id="error-100"></a>Error code 100 - Server error

This error is returned when something went wrong on the server-side. HTTP response code 500 (Internal server error)
