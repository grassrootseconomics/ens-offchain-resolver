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
- [x] Read multicoin address (Celo)
- [ ] Read content hash
- [ ] Read text record

### Integration guide

To register names:

If the name is available, registeration will be done immidiately, otherwise a
random name will be choosen (upto 90 iterations) based on our naive autoChoose
algo.

```bash
> POST http://localhost:5015/api/v1/internal/register
> authorization: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJwdWJsaWNLZXkiOiIweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAiLCJzZXJ2aWNlIjp0cnVlLCJpc3MiOiJldGgtY3VzdG9kaWFsLWRldiIsInN1YiI6InNuLXByb2QiLCJleHAiOjE3Njk4NDIxNzksImlhdCI6MTczODMwNjE3OX0.FXTwZ8nQKCG66xO0wMbx4Mga8SqFZcm65pq7_iMKjXPMH_h0IBHmSV2DOKQVfNbI1W9BRUCuSUwbALFgDqLrBg
> content-type: application/json
> data {"address":"0xF7D1D901d15BBf60a8e896fbA7BBD4AB4C1021b3","hint":"peterxd.sarafu.eth"}
```

response:

```json
{
    "ok": true,
    "description": "Name registered",
    "result": {
        "address": "0xF7D1D901d15BBf60a8e896fbA7BBD4AB4C1021b3",
        "autoChoose": true,
        "name": "peterxd71.sarafu.eth"
    }
}
```

To resolve names (name to address):

```bash
> GET http://localhost:5015/api/v1/resolve/peterxd71.sarafu.eth
> authorization: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJwdWJsaWNLZXkiOiIweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAiLCJzZXJ2aWNlIjp0cnVlLCJpc3MiOiJldGgtY3VzdG9kaWFsLWRldiIsInN1YiI6InNuLXByb2QiLCJleHAiOjE3Njk4NDIxNzksImlhdCI6MTczODMwNjE3OX0.FXTwZ8nQKCG66xO0wMbx4Mga8SqFZcm65pq7_iMKjXPMH_h0IBHmSV2DOKQVfNbI1W9BRUCuSUwbALFgDqLrBg
> content-type: application/json
```

response:

```json
{
    "ok": true,
    "description": "Address resolved",
    "result": {
        "address": "0xF7D1D901d15BBf60a8e896fbA7BBD4AB4C1021b3"
    }
}
```

To reverse resolve (address to name):

```bash
> GET http://localhost:5015/api/v1/resolve/reverse/0xF7D1D901d15BBf60a8e896fbA7BBD4AB4C1021b3
> authorization: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJwdWJsaWNLZXkiOiIweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAiLCJzZXJ2aWNlIjp0cnVlLCJpc3MiOiJldGgtY3VzdG9kaWFsLWRldiIsInN1YiI6InNuLXByb2QiLCJleHAiOjE3Njk4NDIxNzksImlhdCI6MTczODMwNjE3OX0.FXTwZ8nQKCG66xO0wMbx4Mga8SqFZcm65pq7_iMKjXPMH_h0IBHmSV2DOKQVfNbI1W9BRUCuSUwbALFgDqLrBg
> content-type: application/json
```

response:

```json
{
    "ok": true,
    "description": "Name reverse resolved",
    "result": {
        "name": "peterxd71.sarafu.eth"
    }
}
```

To lookup names:

The resolver supports both Ethereum and Celo address resolutions. For Celo
addresses, use the `addr(bytes32,uint256)` function with the Celo coin type
(2147525868).

Example CCIP read request for Celo address:

```solidity
// In your contract
bytes memory data = abi.encodeWithSelector(
    bytes4(keccak256("addr(bytes32,uint256)")),
    namehash,
    2147525868 // Celo coin type
);
```

## License

[AGPL-3.0](LICENSE).
