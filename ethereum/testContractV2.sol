pragma solidity ^0.4.0;
contract FactomAnchor {

    struct Anchor {
        uint256 KeyMR;
    }

    /*Public accessors!*/
    /*http://solidity.readthedocs.io/en/latest/contracts.html?highlight=accessor#accessor-functions*/
    
    address public creator;
    mapping(uint256 => Anchor) public anchors;
    string[] public debug;

    /*********************************Events********************************/
    event Debug(string info);
    event AnchorMade(uint256 height, uint256 merkleroot);

        //Contract initialization
        function FactomAnchor() {
            creator = msg.sender;
        }
        
        /*******************************Modifiers*******************************/

        modifier onlyCreator {
            //only creator can perform some actions until it disables itself
            require(msg.sender == creator);
            _;
        }
        
        function debugEntry(string message) {
            var id = debug.length++;
            debug[id] = message;
            Debug(message);
        }
        
        /*******************************Functions*******************************/
        //Set Factom anchors
        function setAnchor(uint256 blockNumber, uint256 keyMR) onlyCreator {
            anchors[blockNumber].KeyMR = keyMR;
            AnchorMade(blockNumber, keyMR);
        }
}
