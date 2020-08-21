//
// Copyright (c) 2020 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/command"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"

	"github.com/student3671/app-functions-sdk-go/internal/bootstrap/container"
	"github.com/student3671/app-functions-sdk-go/internal/common"
)

// Clients contains references to dependencies required by the Clients bootstrap implementation.
type Clients struct {
}

// NewClients create a new instance of Clients
func NewClients() *Clients {
	return &Clients{}
}

// BootstrapHandler setups all the clients that have be specified in the configuration
func (_ *Clients) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	config := container.ConfigurationFrom(dic.Get)

	var eventClient coredata.EventClient
	var valueDescriptorClient coredata.ValueDescriptorClient
	var commandClient command.CommandClient
	var notificationsClient notifications.NotificationsClient

	// Use of these client interfaces is optional, so they are not required to be configured. For instance if not
	// sending commands, then don't need to have the Command client in the configuration.
	if _, ok := config.Clients[common.CoreDataClientName]; ok {
		eventClient = coredata.NewEventClient(
			local.New(config.Clients[common.CoreDataClientName].Url() + clients.ApiEventRoute))

		valueDescriptorClient = coredata.NewValueDescriptorClient(
			local.New(config.Clients[common.CoreDataClientName].Url() + clients.ApiValueDescriptorRoute))
	}

	if _, ok := config.Clients[common.CoreCommandClientName]; ok {
		commandClient = command.NewCommandClient(
			local.New(config.Clients[common.CoreCommandClientName].Url() + clients.ApiDeviceRoute))
	}

	if _, ok := config.Clients[common.NotificationsClientName]; ok {
		notificationsClient = notifications.NewNotificationsClient(
			local.New(config.Clients[common.NotificationsClientName].Url() + clients.ApiNotificationRoute))
	}

	// Note that all the clients are optional so some or all these clients may be nil
	// Code that uses them must verify the client was defined and created prior to using it.
	// This information is provided in the documentation.
	dic.Update(di.ServiceConstructorMap{
		container.EventClientName: func(get di.Get) interface{} {
			return eventClient
		},
		container.ValueDescriptorClientName: func(get di.Get) interface{} {
			return valueDescriptorClient
		},
		container.CommandClientName: func(get di.Get) interface{} {
			return commandClient
		},
		container.NotificationsClientName: func(get di.Get) interface{} {
			return notificationsClient
		},
	})

	return true
}
