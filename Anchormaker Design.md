Anchormaker Design
=============

Anchormaker is a standalone application that is designed to run alongside factomd, bitcoind and geth to anchor Factom Directory Blocks into Bitcoin and Ethereum.

Design questions
--------

1) Should we stall to save an anchor in a Factoid Entry for both Bitcoin and Ethereum together if possible, or race to anchor a record as fast as possible? Since we're anchored ASAP anyway, waiting to save anchor entry isn't too much of a priority and it would produce overall less entries, but on the flip-side we could stall the anchors for awhile if one network is being unresponsive... For now anchoring together if possible.

General functionality
--------

The main Anchormaker directory is responsible for loading the configuration and the database file, setting up the three networks and running them on loop.

To ensure the program is safely shut down, interrupts only happen after a full loop. There is a separate goroutine for hard interrupts that disregard safety.

The program is not designed for multiple threads (other than the interrupt thread). This is done intentionally, as the anchoring process has a strict ordering (see below).

The configuration for all of the networks is read by the main loop of the program and sent to a LoadConfig function for each of the network before the main loop is started.

Anchormaker uses a single, shared database. It is passed by the main loop of the program down to every function that needs it.

Database and data
--------

The database used by Anchormaker is an expansion of factomd's DatabaseOverlay. It can easily support Bolt and LevelDB. It has been expanded with functions specific to Anchormaker and its data.

The main piece of data stored by Anchormaker is AnchorData. Every entry represents a single Directory Block and all of its anchors. It is indexed by DB's Height.

The database also stores ProgramState as an indicator of how far Anchormaker is synchronized with the various networks.

Anchoring order
--------

Anchormaker has a strict order for anchoring data: 

1) When a new Directory Block is first created, its AnchorData will only have KeyMR and Height.

2) When it is anchored into Bitcoin or Ethereum, it will have the network's TxID, Address, etc. field populated.

3) When the anchor is confirmed and embedded into the blockchain permanently, the record is updated with the block hash and height the transaction was included in.

4) After the confirmed anchor is saved, it can be embedded into Factom's entry chain. When an entry is sent out to be saved, the Entry Hash is saved in AnchorData.

5) When an entry is embedded in a full DirectoryBlock, we save its Record Height (which DBlock height it was included in). This indicates that the anchor cycle for the given network is fully finished.

6) When both entries are fully embedded into DirectoryBlock, the AnchorData is considered complete and can be ignored from now on.

Synchronization process
--------

Before any new anchor is allowed to be created, Anchormaker needs to fully synchronize itself with all of the networks. This is to ensure we're not double-anchoring the same pieces of data multiple times.

All of data synchronization is run in a loop until no new data is found in a whole loop (no new DBlocks, Bitcoin or Ethereum Transactions). This is to ensure all data is up-to-date in case some part of the synchronization took too long.

First, the program synchronizes Factom's Directory Blocks and Anchor Entries. This is to ensure we know about every DBlock and all of the existing anchors.

Afterwards, Bitcoin and Ethereum transactions are synchronized to fetch any anchors that are not yet embedded into Factom.

The synchronization process should be mindful of melleated transactions and update the TxIDs accordingly.

Only once all anchors are fetched from all three networks, should new anchors be created.

Anchor locations
--------

Every network stores anchors in a specific location to ensure it is easy to find them. In Factom, the anchors are stored in chain df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604 (? different chain for Ethereum ?). In Bitcoin, the anchors are issued by address ??? (? more than one address ?). In Ethereum, all anchors are transactions from wallet address 0x838f9b4d8ea3ff2f1bd87b13684f59c4c57a618b and smart contract address 0x8a8fbabbec1e99148083e9314dffd82395dd8f18 (? TODO: change for production data ?). In the future more anchor locations can be added, but for now those are the only places Anchormaker looks at when trying to find anchors.

Creating new anchors
--------

Once all data is synchronized with all of the networks, Anchormaker can start making new anchors. It should first check its wallet balance to ensure it has enough funds to proceed with creating anchors. Afterwards, it can create new anchors in the following way:

1) Every network anchoring loop should start at the last complete AnchorData (having both Bitcoin and Ethereum anchor's RecordHeight from Factom) and iterate onwards from there.

2) The loop should skip over entries it already tried anchoring (AnchorData will have their TxIDs)

3) For every DirectoryBlock that doesn't have the network-specific anchor transaction, the loop should create a new transaction, send it to the network, and upon success - immediately record the TxID into AnchorData.

4) The transactions should have sufficient fees and never try to double-spend themselves or spend unconfirmed transactions to ensure no transaction gets orphaned or remains unconfirmed.

5) Continue until all DirectoryBlocks have an anchor transaction.