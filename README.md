[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/sigstore/rekor/badge)](https://api.securityscorecards.dev/projects/github.com/sigstore/rekor)

<p align="center">
  <img style="max-width: 100%;width: 300px;" src="./images/logo.svg" alt="Rekor logo"/>
</p>

# Rekor

Rekór - Greek for “Record”

Rekor's goals are to provide an immutable tamper resistant ledger of metadata generated within a software projects supply chain.
Rekor will enable software maintainers and build systems to record signed metadata to an immutable record.
Other parties can then query said metadata to enable them to make informed decisions on trust and non-repudiation of an object's lifecycle. For more details visit the [sigstore website](https://sigstore.dev)

The Rekor project provides a restful API based server for validation and a transparency log for storage.
A CLI application is available to make and verify entries, query the transparency log for inclusion proof,
integrity verification of the transparency log or retrieval of entries by either public key or artifact.

Rekor fulfils the signature transparency role of sigstore's software signing
infrastructure. However, Rekor can be run on its own and is designed to be
extensible to working with different manifest schemas and PKI tooling.

[Official Documentation](https://docs.sigstore.dev/rekor/overview).

## Public Instance

We're currently working hard on cutting a 1.0 release and productionizing the public instance.
We don't have a date yet, but follow along on the [GitHub project](https://github.com/orgs/sigstore/projects/5).

We will improve the stability and publish SLOs soon!

More details on the public instance can be found at [docs.sigstore.dev](https://docs.sigstore.dev/rekor/public-instance).

### Installation

Please see the [installation](https://docs.sigstore.dev/rekor/overview#usage-and-installation) page for details on how to install the rekor CLI and set up / run
the rekor server

### Usage

For examples of uploading signatures for all the supported types to rekor, see [the types documentation](types.md).

## Extensibility

### Custom schemas / manifests (rekor type)

Rekor allows customized manifests (which term them as types), [type customization is outlined here](https://github.com/sigstore/rekor/tree/main/pkg/types).

### API

If you're interesting in integration with Rekor, we have an [OpenAPI swagger editor](https://sigstore.dev/swagger/)

## Security

Should you discover any security issues, please refer to sigstore's [security process](https://github.com/sigstore/.github/blob/main/SECURITY.md)

## Contributions

We welcome contributions from anyone and are especially interested to hear from
users of Rekor.
