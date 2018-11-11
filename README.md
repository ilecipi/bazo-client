# bazo-client

The command line interface for interacting with the Bazo blockchain implemented in Go.

## Setup Instructions

The programming language Go (developed and tested with version >= 1.9) must be installed, the properties `$GOROOT` and `$GOPATH` must be set. 
For more information, please check out the [official documentation](https://github.com/golang/go/wiki/SettingGOPATH).

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
* `--rootkey`: Load root's private key from this file
* `--file`: Save the new account's public and private key to this file

Examples

```bash
bazo-client account create --rootkey root.txt --file newaccount.txt
bazo-client account create --rootkey root.txt --file newaccount.txt --fee 5
```

#### Add Account

Add an existing account to the network.

```bash
bazo-client account add [command options] [arguments...]
```

Options
* `--header`: (default: 0) Set header flag
* `--fee`: (default: 1) Set transaction fee
* `--rootkey`: Load root's private key from this file
* `--address`: Existing account's 128 byte address

```bash
bazo-client account create --rootkey root.txt --address b978...<120 byte omitted>...e86ba
bazo-client account create --rootkey root.txt --address b978...<120 byte omitted>...e86ba --fee 5 
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
* `--toFile`: The file to load the recipient's address from
* `--toAddress`: Instead of passing the recipient's address by file with `--toFile`, you can also directly pass the recipient's address
* `--multisig`: (optional) The file to load the multisig's private key from

Examples

```bash
bazo-client funds --from myaccount.txt --txcount 0 --toFile recipient.txt --amount 100
bazo-client funds --from myaccount.txt --txcount 1 --toFile recipient.txt --amount 100 --multisig multisig.txt
bazo-client funds --from myaccount.txt --txcount 2 --toAddress b978...<120 byte omitted>...e86ba --amount 100 --fee 15
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
* `--rootkey`: Load root's private key from this file
* `--setBlockSize`: Set the size of blocks (in bytes)
* `--setDifficultyInterval`: Set the difficulty interval (in number of blocks) 
* `--setMinimumFee`: Set the minimum fee (in Bazo coins)
* `--setBlockInterval`: Set the block interval (in seconds)
* `--setBlockReward`: Set the block reward (in Bazo coins)

Examples

```bash
bazo-client network --txcount 0 --rootkey root.txt --setBlockSize 2048
bazo-client network --txcount 1 --rootkey root.txt --setDifficultyInterval 10
bazo-client network --txcount 2 --rootkey root.txt --setMinimumFee 10
bazo-client network --txcount 3 --rootkey root.txt --setBlockInterval 120
bazo-client network --txcount 4 --rootkey root.txt --setBlockReward 5
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
* `--key`: The file to load the validator's private key from
 
#### Enable Staking
 
Join the pool of validators.
 
```bash
bazo-client staking enable [command options] [arguments...]
```

Options
* `--commitment`: The file to load the validator's commitment key from (will be created if it does not exist)

Example

```bash
bazo-client staking enable --key myaccount.txt --commitment commitment.txt
```
 
 #### Disable Staking
 
Leave the pool of validators.
 
```bash
bazo-client staking disable [command options] [arguments...]
```

Example

```bash
bazo-client staking disable --key myaccount.txt
```

### REST 

Start the REST service.

```bash
bazo-client rest
```
