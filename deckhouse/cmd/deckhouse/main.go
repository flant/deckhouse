package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/romana/rlog"
	"gopkg.in/alecthomas/kingpin.v2"

	shell_operator_app "github.com/flant/shell-operator/pkg/app"
	"github.com/flant/shell-operator/pkg/executor"
	utils_signal "github.com/flant/shell-operator/pkg/utils/signal"

	addon_operator_app "github.com/flant/addon-operator/pkg/app"

	"flant/deckhouse/pkg/app"
	"flant/deckhouse/pkg/deckhouse"
)

// Variables with component versions. They set by 'go build' command.
var DeckhouseVersion = "dev"
var AddonOperatorVersion = "dev"
var ShellOperatorVersion = "dev"

func main() {
	// TODO DELETE THIS AFTER MIGRATION
	// temporary fix to migrate from ANTIOPA_POD to DECKHOUSE_POD
	antiopaPod := os.Getenv("ANTIOPA_POD")
	if antiopaPod != "" {
		err := os.Setenv("DECKHOUSE_POD", antiopaPod)
		if err != nil {
			panic(err)
		}
	}
	// END DELETE THIS AFTER MIGRATION

	shell_operator_app.Version = ShellOperatorVersion
	addon_operator_app.Version = AddonOperatorVersion

	rlog.Infof("deckhouse %s (shell-operator %s, addon-operator %s)", DeckhouseVersion, ShellOperatorVersion, AddonOperatorVersion)

	kpApp := kingpin.New(app.AppName, fmt.Sprintf("%s %s: %s", app.AppName, DeckhouseVersion, app.AppDescription))

	// global defaults
	app.SetupGlobalSettings(kpApp)
	// set global options for addon-operator
	addon_operator_app.SetupGlobalSettings(kpApp)

	// start main loop
	kpApp.Command("start", "Start deckhouse.").
		Default().
		Action(func(c *kingpin.ParseContext) error {
			// Be a good parent - clean up after the child processes
			// in case if addon-operator is a PID 1 process.
			go executor.Reap()

			deckhouse.Start()

			// Block action by waiting signals from OS.
			utils_signal.WaitForProcessInterruption()

			return nil
		})

	kingpin.MustParse(kpApp.Parse(os.Args[1:]))

	return
}