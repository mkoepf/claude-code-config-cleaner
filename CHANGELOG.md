# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Claude Code Config Cleaner (`cccc`)
- `clean projects` command to remove stale project session data
- `clean orphans` command to remove orphaned data (empty sessions, orphan todos, file-history)
- `clean config` command to deduplicate local configs against global settings
- `list projects` command with `--stale-only` filter
- `list orphans` command to preview orphaned data
- `list config` command with `--verbose` mode
- Dry-run support (`--dry-run`) for all clean operations
- Audit logging to `~/.claude/cccc-audit.log`
- Safe by default: all destructive operations require confirmation (`-y` to skip)
- CI pipeline with coverage and security scanning
