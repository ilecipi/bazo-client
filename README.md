# bazo-client

The command line interface for interacting with the Bazo blockchain implemented in Go.

## Setup Instructions

The programming language Go (developed and tested with version >= 1.9) must be installed, the properties `$GOROOT` and `$GOPATH` must be set. 
For more information, please check out the [official documentation](https://github.com/golang/go/wiki/SettingGOPATH).

Furthermore, the `configuration.json` in the root directory of the bazo-client must be properly configured before 
interacting with the CLI.

Contents of `configuration.json`:
```json
{
  "this_client": {
    "ip": "127.0.0.1",
    "port": "8010"
  },
  "bootstrap_server": {
    "ip": "127.0.0.1",
    "port": "8000"
  },
  "multisig_server": {
    "ip": "127.0.0.1",
    "port": "8020"
  }
}
```

## Getting Started

The Bazo client provides an intuitive and beginner-friendly command line interface.

```bash
bazo-client [global options] command [command options] [arguments...]
```

Options
* `--help, -h`: Show help 
* `--version, -v`: Print the version

### Accounts

While everybody can check the state of accounts, only somebody in possession of the root private key can
create accounts or add existing accounts to the network.

#### Check Account State

```bash
bazo-client account check [command options] [arguments...]
```

Options
* `--file`: Load the 128 byte address from a file
* `--address`: Instead of passing the account's address by file with `--file`, you can also directly pass the 128 byte address

Examples

```bash
bazo-client account check b978...<120 byte omitted>...e86ba
bazo-client account check myaccount.txt 
```

#### Create Account

Create a new account and add it to the network. Save the public-private keypair to a file.

```bash
bazo-client account create [command options] [arguments...]
```

Options
* `--header`: (default: 0) Set header flag
* `--fee`: (default: 1) Set transaction fee
* `--rootwallet`: Load root's private key from this file
* `--file`: Save the new account's public and private key to this file

Examples

```bash
bazo-client account create --rootwallet root.txt --wallet newaccount.txt
bazo-client account create --rootwallet root.txt --wallet newaccount.txt --fee 5
```

#### Add Account

Add an existing account to the network.

```bash
bazo-client account add [command options] [arguments...]
```

Options
* `--header`: (default: 0) Set header flag
* `--fee`: (default: 1) Set transaction fee
* `--rootwallet`: Load root's private key from this file
* `--address`: Existing account's 128 byte address

```bash
bazo-client account create --rootwallet root.txt --address b978...<120 byte omitted>...e86ba
bazo-client account create --rootwallet root.txt --address b978...<120 byte omitted>...e86ba --fee 5 
```

### Funds

Send Bazo coins from one account to another.

```bash
bazo-client funds [command options] [arguments...]
```

Options
* `--header`: (default: 0) Set header flag
* `--fee`: (default: 1) Set transaction fee
* `--txcount`: The sender's current transaction counter
* `--amount`: The amount to transfer from sender to recipient
* `--from`: The file to load the sender's private key from
* `--to`: The file to load the recipient's public key from
* `--toAddress`: Instead of passing the recipient's address by file with `--to`, you can also directly pass the recipient's address with this option
* `--multisig`: (optional) The file to load the multisig's private key from.

Examples

```bash
bazo-client funds --from myaccount.txt --to recipient.txt --txcount 0 --amount 100
bazo-client funds --from myaccount.txt --to recipient.txt --txcount 1 --amount 100 --multisig myaccount.txt
bazo-client funds --from myaccount.txt --toAddress b978...<120 byte omitted>...e86ba --txcount 2 --amount 100 --fee 15
```

### Network

Configure network settings.

```bash
bazo-client network [command options] [arguments...]
```

Options
* `--header`: (default: 0) Set header flag
* `--fee`: (default: 1) Set transaction fee
* `--txcount`: The sender's current transaction counter
* `--rootwallet`: Load root's private key from this file
* `--setBlockSize`: Set the size of blocks (in bytes)
* `--setDifficultyInterval`: Set the difficulty interval (in number of blocks) 
* `--setMinimumFee`: Set the minimum fee (in Bazo coins)
* `--setBlockInterval`: Set the block interval (in seconds)
* `--setBlockReward`: Set the block reward (in Bazo coins)

Examples

```bash
bazo-client network --txcount 0 --rootwallet root.txt --setBlockSize 2048
bazo-client network --txcount 1 --rootwallet root.txt --setDifficultyInterval 10
bazo-client network --txcount 2 --rootwallet root.txt --setMinimumFee 10
bazo-client network --txcount 3 --rootwallet root.txt --setBlockInterval 120
bazo-client network --txcount 4 --rootwallet root.txt --setBlockReward 5
```

Note that each setting broadcasts one `ConfigTx` to the network.

### Staking

Join or leave the pool of validators by enabling or disabling staking.

```bash
 bazo-client staking [command options] [arguments...]
```

Options: 
* `--header`: (default: 0) Set header flag
* `--fee`: (default: 1) Set transaction fee
* `--wallet`: The file to load the validator's private key from
 
#### Enable Staking
 
Join the pool of validators.
 
```bash
bazo-client staking enable [command options] [arguments...]
```

Options
* `--commitment`: The file to load the validator's commitment key from. A new commitment key is generated if it does not exist yet.

Example

```bash
bazo-client staking enable --wallet mywallet.txt --commitment commitment.txt
```
 
 #### Disable Staking
 
Leave the pool of validators.
 
```bash
bazo-client staking disable [command options] [arguments...]
```

Example

```bash
bazo-client staking disable --wallet myaccount.txt
```

### REST 

Start the REST service.

```bash
bazo-client rest
```
