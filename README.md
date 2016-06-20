Anchormaker
=============

Setup Instructions
--------

First, you must install btcd and btcwallet. To do this, follow the "Install BTCD" and "Install Wallet" instructions on [this page.](https://github.com/FactomProject/FactomDocs/blob/daaf437d59a31db0d5ed4ccf9eadb6205bc5da6b/factomFullInstall.md#install-btc-suite)

Next, you will want to download anchormaker and copy the anchormaker.conf file to your .factom directory:

```
go get -v github.com/FactomProject/anchormaker/...
cp $GOPATH/src/github.com/FactomProject/anchormaker/anchormaker.conf $HOME/.factom/anchormaker.conf
```

Finally, in order for anchormaker to be able to make entries into Factom, you must modify the $HOME/.factom/anchormaker.conf file and change the "ServerECKey" value from "e1" to a named entry credit address in your Factom wallet.


Running Anchormaker
--------

First, make sure that both factomd and fctwallet are running.


**Note: Because this program will automatically write anchor records to the anchor chain [df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604](http://explorer.factom.org/chain/df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604), it is recommended that you run this program on a Factom sandbox, rather than on the live mainnet.**

If you are doing this in a sandbox, the anchor chain will need to be created before the program is first run:

```
echo "This is the Factom anchor chain, which records the anchors Factom puts on Bitcoin and other networks." | factom-cli mkchain -e FactomAnchorChain e1
```

(The example above assumes you have an entry credit address in your wallet named "e1")

Next, make sure that btcd and btcwallet are also running.

In a new terminal window, after btcd and btcwallet are fully synchronized with the Bitcoin blockchain, run:

```
btcctl --rpcuser=testuser --rpcpass=SecurePassHere --testnet --wallet getaccountaddress default
```

This will output a Bitcoin address. Visit [this page](https://testnet.manu.backend.hamburg/faucet) and enter the address in the bar next to the "Give me some coins" button. Click that button a few times to send testnet coins to your Bitcoin wallet.

You will need to wait a few minutes for the coins to be confirmed in your wallet. You can test if the transaction(s) are confirmed with this command:

```
btcctl --rpcuser=testuser --rpcpass=SecurePassHere --testnet --wallet getbalance default
```

Once this balance is non-zero, you are able to run anchormaker successfully. From the $HOME/github.com/FactomProject/anchormaker/ folder, you can run:

```
go build
./anchormaker
```
