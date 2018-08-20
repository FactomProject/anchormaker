pragma solidity ^0.4.24;
contract FactomAnchor {
    
    struct Anchor {
        uint256 MerkleRoot;
    }

    //****************************Public accessors***************************
    
    address public creator;
    mapping(uint256 => Anchor) public anchors;
    bool public frozen;

    //*********************************Events********************************
    event AnchorMade(uint256 height, uint256 merkleRoot);
    event AnchoringFrozen(uint256 height);  

    //Contract initialization
    constructor() public {
        creator = msg.sender;
        frozen = false;
    }

    //*******************************Modifiers*******************************

    modifier onlyCreator {
        //only creator can perform some actions until it disables itself
        require(msg.sender == creator);
        _;
    }

    //*******************************Functions*******************************
    //Set Factom anchors
    function setAnchor(uint256 blockNumber, uint256 merkleRoot) public onlyCreator {
        if (!frozen) {
            anchors[blockNumber].MerkleRoot = merkleRoot;
            emit AnchorMade(blockNumber, merkleRoot);
        }
    }
    
    //Get Factom anchors
    function getAnchor(uint256 blockNumber) public constant returns (uint256) {
        return anchors[blockNumber].MerkleRoot;
    }
    
    //stop future updates
    function freeze(uint256 height) public onlyCreator {
        frozen = true;
        emit AnchoringFrozen(height);
    }
    
    //checks if state is stopped from being updated later
    function checkFrozen() public constant returns (bool) {
        return frozen;
    }
       
}
