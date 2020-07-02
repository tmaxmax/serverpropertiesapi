# Changelog

## [1.2.2](https://github.com/tmaxmax/serverpropertiesapi/compare/1.2.1...1.2.2)

### Changed

- API endpoints make more sense
- API is served on https

### Fixed

- Unindented marshaled JSON from single property API response is now indented
- Accepted content-type checking now works for multiple accepted types

## [1.2.1](https://github.com/tmaxmax/serverpropertiesapi/compare/1.2.0...1.2.1)

### Added

- Filter values can now be comma-separated
- The responses are sent gzipped if the client supports it

### Changed

- Marshaled JSON is now indented

### Fixed

- Wrong header key for Content-Type fixed, was Content-Types
- Filter values are now validated, the API returning a 400 Bad Request error when they are invalid.

## [1.2.0](https://github.com/tmaxmax/serverpropertiesapi/compare/1.1.0...1.2.0)

### Added

- API endpoint for requesting from the whole documentation now supports options to filter the results (see [README.md](README.md))

### Changed

- API now scrapes the website on each request, if not cached (this ensures the information is always up to date)
- Requests are cached for 24 hours

## [1.1.0](https://github.com/tmaxmax/serverpropertiesapi/compare/1.0.1...1.1.0)

### Added

- API endpoint for requesting metadata, such as property types and limit default value
- Unsupported HTTP methods now have their own handler function

## [1.0.1](https://github.com/tmaxmax/serverpropertiesapi/compare/1.0.0...1.0.1)

### Added

- The project now has a changelog

### Changed

- Scraper ignores possible values for boolean properties, as they are redundant

### Fixed

- Scraper now picks up the possible values correctly

## [1.0.0](https://github.com/tmaxmax/serverpropertiesapi/releases/tag/1.0.0)

This is the first API release

### Added

- Implemented the Minecraft Gamepedia Wiki server.properties page scraper
- API endpoint for requesting from whole documentation
- API endpoint for requesting the documentation of a single key
- API error handling for requesting unsupported file formats and inexistent resources
