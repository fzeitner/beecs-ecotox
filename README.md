# beecs-ecotox

[![Test status](https://img.shields.io/github/actions/workflow/status/mlange-42/beecs/tests.yml?branch=main&label=Tests&logo=github)](https://github.com/mlange-42/beecs/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mlange-42/beecs)](https://goreportcard.com/report/github.com/fzeitner/beecs-ecotox)
[![GitHub](https://img.shields.io/badge/github-repo-blue?logo=github)](https://github.com/fzeitner/beecs-ecotox)

Re-implementation of the [BEEHAVE_ecotox](https://academic.oup.com/etc/article/41/11/2870/7730717) model
in [Go](https://go.dev) using the [Ark](https://github.com/mlange-42/ark) Entity Component System (ECS).
This repository encompasses only the most essential parts of BEEHAVE_ecotox and added these to the already established
[beecs](https://github.com/mlange-42/beecs) model.

All the hard work to develop, parameterize and validate the original BEEHAVE model was done by Dr. Matthias Becher and co-workers.
I was not involved in that development in any way, and just re-implement the model following its ODD Protocol and the NetLogo code.
The ecotoxicological submodels were developed by Thomas G. Preuss, Benoit Goussen, Matthias Becher and their co-workers.
I only re-implemented these into the beecs model framework.

Beecs-ecotox currently implements BEEHAVE's basic colony and foraging models. Only basic ecotoxicological model simulations are possible at the moment; the newer options, like simulating multiple different pesticide applications within one simulation run, from a later amendment to the NetLogo model (see https://github.com/ibacon-GmbH-Modelling/BEEHAVEecotox) are not yet added.
Colony dynamics, exposure, and dose-response mimic BEEHAVE_ecotox almost perfectly. See the [tests](https://github.com/fzeitner/beecs-ecotox/tree/main/_tests) for a comparison between both implementations.

## Usage

[beecs-cli](https://github.com/mlange-42/beecs-cli) and [beecs-ui](https://github.com/mlange-42/beecs-ui) have not yet been tested for compatibility to this repository.

The easiest way to use this model is to simply create a local copy and run it using main-files. To create a local copy of this repository simply run
```
git clone https://github.com/fzeitner/beecs-ecotox

```

Usage examples can be found in folder [_examples](https://github.com/mlange-42/beecs/blob/main/_examples).

## Exploring the model

To explore the model code, start with reading files [`model/default.go`](https://github.com/mlange-42/beecs/blob/main/model/default.go)
and [`params/default.go`](https://github.com/mlange-42/beecs/blob/main/params/default.go).
Given that the model is developed with ECS, the structure should be quite obvious.

- [`comp`](https://github.com/mlange-42/beecs/blob/main/comp) contains all components.
- [`sys`](https://github.com/mlange-42/beecs/blob/main/sys) contains all systems.
- [`params`](https://github.com/mlange-42/beecs/blob/main/params) contains model parameters (as ECS resources).
- [`globals`](https://github.com/mlange-42/beecs/blob/main/globals) contains global variables (as ECS resources).

Also check out the [examples](https://github.com/mlange-42/beecs/blob/main/_examples).
