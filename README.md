[![ci](https://github.com/piprate/sequel-flow-contracts/actions/workflows/ci.yml/badge.svg)](https://github.com/piprate/sequel-flow-contracts/actions/workflows/ci.yml)

# Flow contracts for Sequel

This repository contains the smart contracts and supporting Go framework for the [Flow](https://www.docs.onflow.org)
blockchain used by Sequel marketplace.

The smart contracts are written in [Cadence](https://docs.onflow.org/cadence).

## Contract Addresses

| Contract          | Mainnet              | Testnet              |
|-------------------|----------------------|----------------------|
| Evergreen         | `0x9a02a1d17295f3e7` | `0xfdf325e9204fc94a` |
| DigitalArt        | `0x9a02a1d17295f3e7` | `0xfdf325e9204fc94a` |
| SequelMarketplace | `0x9a02a1d17295f3e7` | `0xfdf325e9204fc94a` |

## Contents

- `contracts/`: All Sequel contracts
- `lib/go/iinft/`: Supporting Go framework
- `lib/go/iinft/scripts`: Useful scripts and transactions made available as Go templates
- `lib/go/iinft/test/`: Test suite for Flow contracts

## About Sequel

Sequel is a new social platform where everything is fun and fictional. It enables you
to be imaginative and express your playful self through sharing creative content
beyond the limits of reality.

See [the official site](https://sequel.space) for more details.

## Gory details

* Flow contracts support the concept of minting a predefined number of editions of a token on demand.
* All tests have transaction fees enabled for extra realism.
