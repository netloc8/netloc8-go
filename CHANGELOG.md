# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `doPost` and `doDelete` internal HTTP transport methods on `Client` for POST and DELETE API calls
- Account management types: `Profile`, `APIKey`, `CreatedKey`, `Usage`, `AuditEntry`, `AuditLog`, `Site`
- Functional option types: `CreateKeyOption` (`WithKeyType`), `AuditLogOption` (`WithLimit`, `WithOffset`, `WithAction`)
- Client methods: `GetProfile`, `ListKeys`, `CreateKey`, `DeleteKey`, `RenewKey`, `GetUsage`, `GetAuditLog`, `ListSites`, `CreateSite`, `DeleteSite`
- Comprehensive test coverage for all account management endpoints (25 tests)
- Package documentation with usage examples for account management, usage statistics, audit log, and allowed origins
