// SPDX-FileCopyrightText: © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package boot_service

import (
	"context"
	"fmt"

	"github.com/OpenCHAMI/ochami/pkg/format"
)

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
