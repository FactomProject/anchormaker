package bitcoin

import (
	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"
)

func LoadConfig(c *config.AnchorConfig) {
	err := InitRPCClient(c)
	if err != nil {
		panic(err)
	}
}

func SynchronizeBitcoinData(dbo *database.AnchorDatabaseOverlay) (int, error) {

	return 0, nil
}

func AnchorBlocksIntoBitcoin(dbo *database.AnchorDatabaseOverlay) error {
	return nil
}
