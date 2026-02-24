// SPDX-FileCopyrightText: © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package boot_service

import (
	"context"
	"fmt"

	boot_service_client "github.com/openchami/boot-service/pkg/client"
	"github.com/openchami/boot-service/pkg/resources/bmc"

	"github.com/OpenCHAMI/ochami/pkg/format"
)

// AddBMCs is a wrapper that calls the boot-service client's CreateBMC()
// function, passing it context. The output is a slice of the BMCs it created,
// each element of which corresponds to an error in an error slice, followed by
// an error that is populatd if an error occurred in the function itself.
func (bsc *BootServiceClient) AddBMCs(token string, bmcs []boot_service_client.CreateBMCRequest) (bmcsAdded []*bmc.BMC, errors []error, funcErr error) {
	// TODO: boot-service client functions don't support tokens yet.
	_ = token

	// TODO: Make concurrent
	for _, bmc := range bmcs {
		ctx, cancel := context.WithTimeout(context.Background(), bsc.Timeout)
		defer cancel()

		item, err := bsc.Client.CreateBMC(ctx, bmc)
		if err != nil {
			newErr := fmt.Errorf("failed to add bmc %+v: %w", bmc, err)
			errors = append(errors, newErr)
			bmcsAdded = append(bmcsAdded, nil)
		}
		bmcsAdded = append(bmcsAdded, item)
	}

	return
}

// GetBMC is a wrapper that calls the boot-service client's GetBMC() function,
// passing it context and a UID. The output is a []byte containing the entity's
// BMC information, formatted as outFormat.
func (bsc *BootServiceClient) GetBMC(token string, outFormat format.DataFormat, uid string) ([]byte, error) {
	// TODO: boot-service client functions don't support tokens yet.
	_ = token

	ctx, cancel := context.WithTimeout(context.Background(), bsc.Timeout)
	defer cancel()

	bcfg, err := bsc.Client.GetBMC(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("request to get BMC info for %s failed: %w", uid, err)
	}

	out, err := format.MarshalData(bcfg, outFormat)
	if err != nil {
		return nil, fmt.Errorf("formatting BMC info for %s failed: %w", uid, err)
	}

	return out, nil
}

// ListBMCs is a wrapper that calls the boot-service client's GetBMCs()
// function, passing it context. The output is a []byte containing a list of
// BMC formatted as outFormat.
func (bsc *BootServiceClient) ListBMCs(token string, outFormat format.DataFormat) ([]byte, error) {
	// TODO: boot-service client functions don't support tokens yet.
	_ = token

	ctx, cancel := context.WithTimeout(context.Background(), bsc.Timeout)
	defer cancel()

	nodes, err := bsc.Client.GetBMCs(ctx)
	if err != nil {
		return nil, fmt.Errorf("request to list BMCs failed: %w", err)
	}

	out, err := format.MarshalData(nodes, outFormat)
	if err != nil {
		return nil, fmt.Errorf("formatting BMC list failed: %w", err)
	}

	return out, nil
}
