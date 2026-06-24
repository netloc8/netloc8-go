# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [1.2.2] — 2026-06-24

### Changed

- Test fixtures in `account_test.go` and `example_test.go` updated to use `location:read` scope (renamed from `geo:read` as part of the monorepo-wide scope migration).

## [1.2.1] — 2026-06-13

### Fixed

- `IsPublicIP` — IPv6 link-local detection now correctly covers the full `fe80::/10` range (`fe80::` through `febf::`). Previously only addresses starting with `fe80` were rejected; addresses like `fe90::1`, `fea0::1`, and `febf::1` were incorrectly treated as public. Now uses byte-level CIDR inspection instead of string prefix matching.

### Added

- Test cases for IPv6 link-local boundary: `fe90::1`, `fea0::1`, `febf::1` (non-public) and `fec0::1` (public, just outside `/10`).

## [1.2.0] — 2026-06-13

### Added

- `RegionCode()` accessor — returns the ISO 3166-2 subdivision code (e.g. `"CA"`), nil-safe
- `Lat()` accessor — returns latitude coordinate, nil-safe
- `Lng()` accessor — returns longitude coordinate, nil-safe
- `HasCoordinates()` predicate — distinguishes absent coordinates from a real `0,0` location

### Changed

- README "Response Shape" section now uses `RegionCode()`, `Lat()`, `Lng()` accessors instead of raw field access

### Fixed

- README nil-safety note now correctly states accessors return zero values (empty strings or `0`), not just "empty strings"

## [1.1.0] — 2026-06-13

### Added

- `doPost` and `doDelete` internal HTTP transport methods on `Client` for POST and DELETE API calls
- Account management types: `Profile`, `APIKey`, `CreatedKey`, `Usage`, `AuditEntry`, `AuditLog`, `Site`
- Functional option types: `CreateKeyOption` (`WithKeyType`), `AuditLogOption` (`WithLimit`, `WithOffset`, `WithAction`)
- Client methods: `GetProfile`, `ListKeys`, `CreateKey`, `DeleteKey`, `RenewKey`, `GetUsage`, `GetAuditLog`, `ListSites`, `CreateSite`, `DeleteSite`
- Comprehensive test coverage for all account management endpoints (25 tests)
- Package documentation with usage examples for account management, usage statistics, audit log, and allowed origins
- Runnable `Example` tests for `CreateKey`, `GetUsage`, `GetAuditLog`
- Account management methods table in README
- `TestNewClient_TimeoutOrdering` test proving option ordering is safe

### Fixed

- `WithTimeout` ordering bug: timeout was silently lost when called before `WithHTTPClient`
- `IsNotFound`, `IsRateLimited`, `IsForbidden` now use `errors.As` instead of direct type assertion, so they work with wrapped errors
- `doPost` no longer sends `Content-Type: application/json` when body is nil (e.g. `RenewKey`)
- Removed false `NETLOC8_API_KEY` env var fallback claim from README (was never implemented)
- `TestClient_CreateKey_NoType` assertion was trivially true; now properly checks the `type` key is absent
