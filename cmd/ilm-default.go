// Copyright (c) 2015-2024 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"fmt"

	"github.com/minio/cli"
	json "github.com/minio/colorjson"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/pkg/v3/console"
)

var ilmDefaultFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "expire-ignore-replication",
		Usage: "expiry ignores replication status, valid values are on, off default is on",
		Value: "on",
	},
}

var ilmDefaultCmd = cli.Command{
	Name:         "default",
	Usage:        "manage lifecycle configuration defaults",
	Action:       mainILMDefault,
	OnUsageError: onUsageError,
	Before:       setGlobalsFromContext,
	Flags:        append(ilmDefaultFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} [FLAGS] TARGET

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
DESCRIPTION:
  Specify default lifecycle expiry behavior when replication is not yet complete.

EXAMPLES:
  1. Avoid expiring eligible objects when replication is not yet complete.
     {{.Prompt}} {{.HelpName}} myminio/mybucket --expire-ignore-replication off

  2. Allow expiring eligible objects when replication is not yet complete.
     {{.Prompt}} {{.HelpName}}  myminio/mybucket --expire-ignore-replication on
`,
}

type ilmDefaultMessage struct {
	Status                  string `json:"status"`
	Target                  string `json:"target"`
	ExpireIgnoreReplication string `json:"expireIgnoreReplication"`
	updated                 bool
}

func (i ilmDefaultMessage) String() string {
	msg := fmt.Sprintf("Lifecycle configuration defaults to expire-ignore-replication: %s", i.ExpireIgnoreReplication)
	if i.updated {
		msg = fmt.Sprintf("Lifecycle configuration defaults updated to expire-ignore-replication: %s for %s", i.ExpireIgnoreReplication, i.Target)
	}
	return console.Colorize(ilmThemeResultSuccess, msg)
}

func (i ilmDefaultMessage) JSON() string {
	msgBytes, e := json.MarshalIndent(i, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")
	return string(msgBytes)
}

func checkILMDefaultSyntax(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		showCommandHelpAndExit(ctx, globalErrorExitStatus)
	}

	waitForRepl := ctx.String("expire-ignore-replication")
	if waitForRepl != "on" && waitForRepl != "off" && waitForRepl != "" {
		fatalIf(errInvalidArgument().Trace(waitForRepl), "valid values are on, off")
	}
}

func mainILMDefault(cliCtx *cli.Context) error {
	ctx, cancelILMDefault := context.WithCancel(globalContext)
	defer cancelILMDefault()

	checkILMDefaultSyntax(cliCtx)
	setILMDisplayColorScheme()
	args := cliCtx.Args()
	urlStr := args.Get(0)

	client, err := newClient(urlStr)
	fatalIf(err.Trace(args...), "Unable to initialize client for "+urlStr+".")

	ilmCfg, _, err := client.GetLifecycle(ctx)
	fatalIf(err.Trace(urlStr), "Unable to fetch lifecycle config, first create a lifecycle config for this bucket")

	if !cliCtx.IsSet("expire-ignore-replication") {
		printMsg(ilmDefaultMessage{
			Status:                  "success",
			Target:                  urlStr,
			ExpireIgnoreReplication: ilmCfg.ExpiryIgnoreReplication,
		})
		return nil
	}
	ilmCfg.ExpiryIgnoreReplication = cliCtx.String("expire-ignore-replication")
	fatalIf(client.SetLifecycle(ctx, ilmCfg).Trace(urlStr), "Unable to set lifecycle rules")

	printMsg(ilmDefaultMessage{
		Status:                  "success",
		Target:                  urlStr,
		updated:                 true,
		ExpireIgnoreReplication: cliCtx.String("expire-ignore-replication"),
	})

	return nil
}
