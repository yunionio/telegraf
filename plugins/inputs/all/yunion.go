package all

import (
	"os"

	execclient "yunion.io/x/executor/client"
	"yunion.io/x/log"

	"github.com/influxdata/telegraf/internal/procutils"

	_ "github.com/influxdata/telegraf/plugins/inputs/ni_rsrc_mon"
	_ "github.com/influxdata/telegraf/plugins/inputs/radeontop"
	_ "github.com/influxdata/telegraf/plugins/inputs/vasmi"
)

const (
	OnecloudExecSockPath = "/hostfs/run/onecloud/exec.sock"
)

func init() {
	log.SetLogLevelByString(log.Logger(), "info")
	if _, err := os.Stat(OnecloudExecSockPath); err == nil {
		log.Infof("init onecloud executor client, socket path: %s", OnecloudExecSockPath)
		execclient.Init(OnecloudExecSockPath)
		execclient.SetTimeoutSeconds(5)
		procutils.SetRemoteExecutor()
	}
}
