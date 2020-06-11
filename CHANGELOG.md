# Changelog

## [1.1.0]

### Added

- API endpoint for requesting metadata, such as property types and limit default value
- Unsupported HTTP methods are now handled

## [1.0.1]

### Added

- The project now has a changelog

### Changed

- Scraper ignores possible values for boolean properties, as they are redundant

### Fixed

- Scraper now picks up the possible values correctly

## [1.0.0]

This is the first API release

### Added

- Implemented the Minecraft Gamepedia Wiki server.properties page scraper
- API endpoint for requesting the whole documentation
- API endpoint for requesting the documentation of a single key
- API error handling for requesting unsupported file formats and inexistent resources
