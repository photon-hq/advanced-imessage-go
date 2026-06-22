# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
While the major version is `0`, the public API may change between minor releases.

## [Unreleased]

### Added

- Initial Go client for the Advanced iMessage API (`photon.imessage.v1`): a thin,
  idiomatic port of `@photon-ai/advanced-imessage`. Covers all eight services
  (addresses, attachments, chats, events, groups, locations, messages, polls),
  range-over-iterator event streams, a structured error type, and a resumable
  subscription helper backed by a caller-supplied `SequenceStore`.
- Minimum supported Go version is **1.24** (CI tests 1.24, 1.25, and 1.26).
