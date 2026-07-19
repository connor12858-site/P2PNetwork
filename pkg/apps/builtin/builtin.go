// Package builtin declares the apps compiled into the desktop client.
package builtin

import (
	"fgov/network/pkg/apps"
	"fgov/network/pkg/apps/testapp"
	textapp "fgov/network/pkg/apps/text"
	"fgov/network/pkg/node"
)

func init() {
	apps.Register(apps.Registration{
		App:      node.App{ID: testapp.ID, Name: testapp.Name, Protocol: testapp.Protocol},
		Register: testapp.Register,
		Run:      testapp.Run,
	})
	apps.Register(apps.Registration{
		App:      node.App{ID: textapp.ID, Name: textapp.Name, Protocol: textapp.Protocol},
		Register: textapp.Register,
		Open:     textapp.Open,
	})
}
