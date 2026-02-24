// SPDX-FileCopyrightText: © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package boot_service

import (
	"context"
	"fmt"

	"github.com/OpenCHAMI/ochami/pkg/format"
)

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
