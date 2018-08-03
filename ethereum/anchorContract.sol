pragma solidity ^0.4.0;
contract FactomAnchor {
    
    struct Anchor {
        uint256 KeyMR;
    }

    //****************************Public accessors***************************
    
    address public creator;
    mapping(uint256 => Anchor) public anchors;
    uint256 public maxHeight;
    bool public frozen;

    //*********************************Events********************************
    event AnchorMade(uint256 height, uint256 merkleroot);
    event AnchorHeightSet(uint256 height);
    event AnchoringFrozen(uint256 height);  

    //Contract initialization
    constructor() public {
        creator = msg.sender;
        maxHeight = 0;
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
    function setAnchor(uint256 blockNumber, uint256 keyMR) public onlyCreator {
        if (!frozen) {
            anchors[blockNumber].KeyMR = keyMR;
            if (blockNumber >= maxHeight) {
                maxHeight = blockNumber;
                emit AnchorHeightSet(maxHeight);
            }
            emit AnchorMade(blockNumber, keyMR);
        }
    }
    
    function setHeight(uint256 newHeight) public onlyCreator {
        if (!frozen) {
            maxHeight = newHeight;
            emit AnchorHeightSet(maxHeight);
        }
    }
    
    //Get Factom anchors
    function getAnchor(uint256 blockNumber) public constant returns (uint256) {
        return anchors[blockNumber].KeyMR;
    }
    
    function getHeight() public constant returns (uint256) {
        return maxHeight;
    }
    
    //stop future updates
    function freeze() public onlyCreator {
        frozen = true;
        emit AnchoringFrozen(maxHeight);
    }
    
    //checks if state is stopped from being updated later
    function checkFrozen() public constant returns (bool) {
        return frozen;
    }
       
}
