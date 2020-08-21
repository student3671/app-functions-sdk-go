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
	"fmt"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/student3671/app-functions-sdk-go/internal/bootstrap/container"
	"github.com/student3671/app-functions-sdk-go/internal/common"
	"github.com/student3671/app-functions-sdk-go/internal/security"
	"github.com/student3671/app-functions-sdk-go/internal/store"
	"github.com/student3671/app-functions-sdk-go/internal/store/db/interfaces"
)

// Database contains references to dependencies required by the database bootstrap implementation.
type Database struct {
}

// NewDatabase create a new instance of Database
func NewDatabase() *Database {
	return &Database{}
}

// BootstrapHandler creates the new interfaces.StoreClient use for database access by Store & Forward capability
func (_ *Database) BootstrapHandler(
	ctx context.Context,
	_ *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	config := container.ConfigurationFrom(dic.Get)

	// Only need the database client if Store and Forward is enabled
	if !config.Writable.StoreAndForward.Enabled {
		dic.Update(di.ServiceConstructorMap{
			container.StoreClientName: func(get di.Get) interface{} {
				return nil
			},
		})
		return true
	}

	logger := bootstrapContainer.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderFrom(dic.Get)

	storeClient, err := InitializeStoreClient(secretProvider, config, startupTimer, logger)
	if err != nil {
		logger.Error(err.Error())
		return false
	}

	dic.Update(di.ServiceConstructorMap{
		container.StoreClientName: func(get di.Get) interface{} {
			return storeClient
		},
	})

	return true
}

// InitializeStoreClient initializes the database client for Store and Forward. This is not a receiver function so that
// it can be called directly when configuration has changed and store and forward has been enabled for the first time
func InitializeStoreClient(
	secretProvider security.SecretProvider,
	config *common.ConfigurationStruct,
	startupTimer startup.Timer,
	logger logger.LoggingClient) (interfaces.StoreClient, error) {
	var err error

	credentials, err := secretProvider.GetDatabaseCredentials(config.Database)
	if err != nil {
		return nil, fmt.Errorf("unable to get Database Credentials for Store and Forward: %s", err.Error())
	}

	config.Database.Username = credentials.Username
	config.Database.Password = credentials.Password

	var storeClient interfaces.StoreClient
	for startupTimer.HasNotElapsed() {
		if storeClient, err = store.NewStoreClient(config.Database); err != nil {
			logger.Warn("unable to initialize Database for Store and Forward: %s", err.Error())
			startupTimer.SleepForInterval()
			continue
		}
		break
	}

	if err != nil {
		return nil, fmt.Errorf("initialize Database for Store and Forward failed: %s", err.Error())
	}

	return storeClient, err
}
