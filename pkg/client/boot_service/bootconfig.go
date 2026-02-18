// SPDX-FileCopyrightText: © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package boot_service

import (
	"context"
	"fmt"

	boot_service_client "github.com/openchami/boot-service/pkg/client"
	"github.com/openchami/boot-service/pkg/resources/bootconfiguration"

	"github.com/OpenCHAMI/ochami/pkg/format"
)

// AddBootConfigs is a wrapper that calls the boot-service client's
// CreateBootConfiguration() function, passing it context. The output is a slice
// of the boot configurations it created, each element of which corresponds to
// an error in an error slice, followed by an error that is populatd if an error
// occurred in the function itself.
func (bsc *BootServiceClient) AddBootConfigs(token string, bootCfgs []boot_service_client.CreateBootConfigurationRequest) (cfgsAdded []*bootconfiguration.BootConfiguration, errors []error, funcErr error) {
	// TODO: boot-service client functions don't support tokens yet.
	_ = token

	// TODO: Make concurrent
	for _, bootCfg := range bootCfgs {
		ctx, cancel := context.WithTimeout(context.Background(), bsc.Timeout)
		defer cancel()

		item, err := bsc.Client.CreateBootConfiguration(ctx, bootCfg)
		if err != nil {
			newErr := fmt.Errorf("failed to add boot configuration %+v: %w", bootCfg, err)
			errors = append(errors, newErr)
			cfgsAdded = append(cfgsAdded, nil)
		}
		cfgsAdded = append(cfgsAdded, item)
	}

	return
}

// GetBootConfig is a wrapper that calls the boot-service client's
// GetBootConfiguration() function, passing it context and a UID. The output is
// a []byte containing the entity's boot configuration, formatted as outFormat.
func (bsc *BootServiceClient) GetBootConfig(token string, outFormat format.DataFormat, uid string) ([]byte, error) {
	// TODO: boot-service client functions don't support tokens yet.
	_ = token

	ctx, cancel := context.WithTimeout(context.Background(), bsc.Timeout)
	defer cancel()

	bcfg, err := bsc.Client.GetBootConfiguration(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("request to get boot configuration for %s failed: %w", uid, err)
	}

	out, err := format.MarshalData(bcfg, outFormat)
	if err != nil {
		return nil, fmt.Errorf("formatting boot configuration for %s failed: %w", uid, err)
	}

	return out, nil
}

// ListBootConfigs is a wrapper that calls the boot-service client's
// GetBootConfigurations() function, passing it context. The output is a []byte
// containing a list of boot configurations formatted as outFormat.
func (bsc *BootServiceClient) ListBootConfigs(token string, outFormat format.DataFormat) ([]byte, error) {
	// TODO: boot-service client functions don't support tokens yet.
	_ = token

	ctx, cancel := context.WithTimeout(context.Background(), bsc.Timeout)
	defer cancel()

	bcfgs, err := bsc.Client.GetBootConfigurations(ctx)
	if err != nil {
		return nil, fmt.Errorf("request to list boot configurations failed: %w", err)
	}

	out, err := format.MarshalData(bcfgs, outFormat)
	if err != nil {
		return nil, fmt.Errorf("formatting boot configuration list failed: %w", err)
	}

	return out, nil
}
