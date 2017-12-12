contract FactomAnchor {

    struct Anchor {
        uint256 KeyMR;
        uint256 Hash;
    }

    /*Public accessors!*/
    /*http://solidity.readthedocs.io/en/latest/contracts.html?highlight=accessor#accessor-functions*/
    
    address public creator;
    mapping(uint256 => Anchor) public anchors;
    string[] public debug;

        //Contract initialization
        function FactomAnchor() {
            creator = msg.sender;
        }
        
        /*******************************Modifiers*******************************/

        modifier onlyCrator {
            //only crator can perform some actions until it disables itself
            if (msg.sender != creator) {
                Debug("Not creator");
                throw;
            }
            Debug("Creator");
            _
        }
        
        /*********************************Events********************************/
        
        
        function Debug(string message) {
            var id = debug.length++;
            debug[id] = message;
        }
        
        /*******************************Functions*******************************/
        //Set Factom anchors
        function setAnchor(uint256 blockNumber, uint256 keyMR, uint256 hash) onlyCrator {
            Debug("setAnchor");
            anchors[blockNumber].KeyMR = keyMR;
            anchors[blockNumber].Hash = hash;
        }
}
