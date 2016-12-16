Anchormaker
=============

Setup Instructions
--------

Download and install BitcoinCore - https://bitcoin.org/en/download .

Next, you will want to download anchormaker and copy the anchormaker.conf file to your .factom directory:

```
go get -v github.com/FactomProject/anchormaker/...
cp $GOPATH/src/github.com/FactomProject/anchormaker/anchormaker.conf $HOME/.factom/anchormaker.conf
```

Finally, in order for anchormaker to be able to make entries into Factom, you must modify the $HOME/.factom/anchormaker.conf file and change the "ServerECKey" value from "e1" to a named entry credit address in your Factom wallet.


Running Anchormaker
--------

First, make sure that factomd is running.


**Note: Because this program will automatically write anchor records to the anchor chain [df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604](http://explorer.factom.org/chain/df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604), it is recommended that you run this program on a Factom sandbox, rather than on the live mainnet.**  chain name: FactomAnchorChain

Next, make sure that bitcoind is also running. 

For testnet, run it with the `-testnet` flag.

First and only the first time you are running bitcoind, make sure to run it with `-txindex=1 -rescan -reindex` flags to make sure we have access to raw transactions. You only need to run it once like this. Make sure bitcoind synchronizes fully with these flags.

Make sure your bitcoin .conf file has the following data (for testnet):

```
testnet=1
server=1
rpcuser=user
rpcpassword=pass
rpcallowip=0.0.0.0/0
rpcport=18332
```

Create a Bitcoin address and put it in your configuration file. Make sure the address has a lot of unspent outputs, roughly 0.1BTC apiece should be good to start.

Once the address balance is non-zero, you are able to run anchormaker successfully. From the $HOME/github.com/FactomProject/anchormaker/ folder, you can run:

```
go build
./anchormaker
```

create new EC addresses in hex and human format [here](https://github.com/FactomProject/Testing/blob/master/examples/python/createECaddress.py)?
