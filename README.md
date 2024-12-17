# ens-offchain-resolver

![GitHub Tag](https://img.shields.io/github/v/tag/grassrootseconomics/ens-offchain-resolver)

Go implementation of the
[EIP-3368 Gateway Interface](https://eips.ethereum.org/EIPS/eip-3668#gateway-interface)
to provide offchain name resolution for `sarafu.eth`. It also includes an
external bypass to resolve without going through the
[CCIP read flow](https://docs.ens.domains/resolvers/ccip-read#ccip-read-flow).

This powers the aliasing system for both Sarafu.Network and USSD.

The resolver deployed to work with this gateway implemntation is the one
provided by
[CCIP tools](https://github.com/ensdomains/ccip-tools/blob/master/contracts/OffchainResolver.sol).

### Supported interfaces

- [x] Read Ethereum address
- [ ] Read multicoin address
- [ ] Read content hash
- [ ] Read text record

## License

[AGPL-3.0](LICENSE).
