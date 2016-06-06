// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/FactomProject/FactomCode/factomlog"
	"github.com/FactomProject/FactomCode/util"
)

var (
	logcfg     = util.ReadConfig().Log
	logPath    = logcfg.LogPath
	logLevel   = logcfg.LogLevel
	logfile, _ = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
)

// setup loggers
var (
	anchorLog = factomlog.New(logfile, logLevel, "ANCHOR")
)
