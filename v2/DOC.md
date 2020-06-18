# Server Properties API documentation

This is MC Bonanza's _server.properties_ reference API documentation. This API can be used to retrieve information from the [official Minecraft Wiki](https://minecraft.gamepedia.com/) about the server.properties keys and values. In this documentation, the term _official documentation_ will refer to the [Minecraft Wiki server.properties documentation](https://minecraft.gamepedia.com/Server.properties), and all its different languages it's written in.

**Root:** `https://api.mcbonanza.games/v2/serverproperties`

### <a id="constants"></a>Constants/Metadata

- <a id="meta-limitNotComputed"></a>**Limit not computed:** `-2147483648` (int32 minimum number)  
  This integral value is used to signify that an integer property has no limits documented in the official documentation. Metadata key: `limitNotComputed`.
- <a id="meta-minecraftTypeNames"></a>**Minecraft boolean type:** `boolean`  
  This string value represents the type-name used to denote integral properties in the official documentation.
- **Minecraft integer type:** `integer`  
  This string value represents the type-name used to denote integral properties in the official documentation.
- **Minecraft string type:** `string`
  This string value represents the type-name used to denote string properties in the official documentation.

## <a id="header"></a>Request header

This section documents the request header keys that modify the response.

- <a id="header-accept"></a>**Accept**  
  The API returns JSON only, meaning that the only accepted MIME-type is
  `application/json`. Including this key isn't necessary, but if it is included and it doesn't contain either `application/json` or `*/*`, the request will be rejected (see the [errors section](#errors)).
- <a id="header-accept-encoding"></a>**Accept-Encoding**  
  The API can return gzipped data. If your client supports it, include `gzip` in this key's value.
- <a id="header-accept-language"></a>**Accept-Language**  
  The API can return the information in the languages that the official website has written the documentation. This should contain only language tags with no subtags, such as `de` or `en` and not `en-US` or `de-CH`. If no language is provided, or if the documentation isn't written in the specified language, English (`en`) will be used as a fallback.

It is not required to add any of these keys to create a valid request. The header can be empty, but these can be used to suit the request to your needs.

## <a id="endpoints"></a>Endpoints

- [GET](#endpoint-allproperties)
- [GET `/{key}`](#endpoint-property)
- [GET `/meta/`](#endpoint-meta)

### <a id="endpoint-allproperties"></a>GET

This endpoint is used to get all properties from the official documentation in the specified language by the [request header](#header-accept-language), or in English, if invalid.

_Example request:_ `GET https://api.mcbonanza.games/v2/serverproperties`  
_Response:_

```json
[]
```

The request can also be filtered queries. Their format can be `key1=val1,val2&key2=val` or `key1=val1&key1=val2&key2=val`.

This is the list of accepted filters:

- **contains** filters the response, so the properties' names contain the passed values. For example, `contains=allow,enable` will filter properties that don't have `allow` and `enable` as substrings, leaving only properties such as `allow-nether` or `enable-rcon`.
- **type** filters the response, so the properties' value types are only the passed values. The only valid types are the type-names mentioned in [constants](#meta-minecraftTypeNames).
- **upcoming** filters the response based on its boolean value. If it's true, the response contains only properties that will be implemented in future versions and mentioned by the official documentation, else it contains only currently implemented properties.

If no element matches the filters, the array will be empty.

### <a id="endpoint-property"></a>GET `/{key}`

This endpoint is used to get a single property named exactly like the `key` parameter's value from the official documentation.

_Example request:_ `GET https://api.mcbonanza.games/v2/serverproperties/allow-nether`
_Response:_

```json
{}
```

### <a id="endpoint-meta"></a>GET `/meta/`

This endpoint is used to get the [constants](#constants) used by the API.

_Example request:_ `GET https://api.mcbonanza.games/v2/serverproperties/meta/`
_Response:_

```json
{}
```

The [language](#header-accept-language) value has no effect on this endpoint.

## <a id="errors"></a>Errors

The API returns the first error encountered in a request in the [RFC 7087 format](https://tools.ietf.org/html/rfc7807#section-3.1).
