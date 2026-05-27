# ZOS SDK Go V4

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/cd6e18aac6be404ab89ec160b4b36671)](https://www.codacy.com/gh/threefoldtech/tfgrid-sdk-go/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=threefoldtech/tfgrid-sdk-go&amp;utm_campaign=Badge_Grade) [![Dependabot](https://badgen.net/badge/Dependabot/enabled/green?icon=dependabot)](https://dependabot.com/) [![Lint](https://github.com/threefoldtech/tfgridv4-sdk-go/actions/workflows/lint.yml/badge.svg?branch=development)](https://github.com/threefoldtech/tfgridv4-sdk-go/actions/workflows/lint.yml)
[![Test](https://github.com/threefoldtech/tfgridv4-sdk-go/actions/workflows/test.yml/badge.svg?branch=development)](https://github.com/threefoldtech/tfgridv4-sdk-go/actions/workflows/test.yml) [![Build](https://github.com/threefoldtech/tfgridv4-sdk-go/actions/workflows/build.yml/badge.svg?branch=development)](https://github.com/threefoldtech/tfgridv4-sdk-go/actions/workflows/build.yml)

Go client library targeting the APIs and features introduced in ThreeFold Grid version 4. It provides updated interfaces for the next-generation grid architecture while maintaining compatibility with existing Go tooling.

## What this is

This repository contains a Go SDK for ThreeFold Grid v4. It offers typed Go interfaces for the new v4 resource types, node registration, and consensus mechanisms. The SDK is intended for Go applications and services that need to interact with the v4 grid APIs.

## What this repository contains

- [node-registrar](./node-registrar/README.md) — Client and utilities for the v4 node registrar API.

## Role in the stack

The SDK operates at the client layer for the v4 grid architecture. It is used by Go-based tools, automation, and backend services that register nodes, query v4 state, or manage v4-specific resources. It complements the broader grid SDK ecosystem by providing Go coverage for the latest grid generation.

## Relation to ThreeFold

This technology is used within the ThreeFold ecosystem and was first deployed on the ThreeFold Grid. The component itself is designed as reusable infrastructure technology and should be understood by its technical function first, independent of any specific deployment.

## Ownership

This repository is owned and maintained by TF-Tech NV, a Belgian company responsible for the development and maintenance of this technology.

## Release

- [release document](./docs/release.md)
- [release script](./release.sh)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
Copyright (c) TF-Tech NV.
