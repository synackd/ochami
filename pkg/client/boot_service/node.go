// SPDX-FileCopyrightText: © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package boot_service

import (
	"context"
	"fmt"

	boot_service_client "github.com/openchami/boot-service/pkg/client"
	"github.com/openchami/boot-service/pkg/resources/node"

	"github.com/OpenCHAMI/ochami/pkg/format"
)

// AddNodes is a wrapper that calls the boot-service client's CreateNode()
// function, passing it context. The output is a slice of the nodes it created,
// each element of which corresponds to an error in an error slice, followed by
// an error that is populatd if an error occurred in the function itself.
func (bsc *BootServiceClient) AddNodes(token string, nodes []boot_service_client.CreateNodeRequest) (nodesAdded []*node.Node, errors []error, funcErr error) {
	// TODO: boot-service client functions don't support tokens yet.
	_ = token

	// TODO: Make concurrent
	for _, node := range nodes {
		ctx, cancel := context.WithTimeout(context.Background(), bsc.Timeout)
		defer cancel()

		item, err := bsc.Client.CreateNode(ctx, node)
		if err != nil {
			newErr := fmt.Errorf("failed to add node %+v: %w", node, err)
			errors = append(errors, newErr)
			nodesAdded = append(nodesAdded, nil)
		}
		nodesAdded = append(nodesAdded, item)
	}

	return
}

// GetNode is a wrapper that calls the boot-service client's GetNode() function,
// passing it context and a UID. The output is a []byte containing the entity's
// node information, formatted as outFormat.
func (bsc *BootServiceClient) GetNode(token string, outFormat format.DataFormat, uid string) ([]byte, error) {
	// TODO: boot-service client functions don't support tokens yet.
	_ = token

	ctx, cancel := context.WithTimeout(context.Background(), bsc.Timeout)
	defer cancel()

	bcfg, err := bsc.Client.GetNode(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("request to get node info for %s failed: %w", uid, err)
	}

	out, err := format.MarshalData(bcfg, outFormat)
	if err != nil {
		return nil, fmt.Errorf("formatting node info for %s failed: %w", uid, err)
	}

	return out, nil
}

// ListNodes is a wrapper that calls the boot-service client's GetNodes()
// function, passing it context. The output is a []byte containing a list of
// nodes formatted as outFormat.
func (bsc *BootServiceClient) ListNodes(token string, outFormat format.DataFormat) ([]byte, error) {
	// TODO: boot-service client functions don't support tokens yet.
	_ = token

	ctx, cancel := context.WithTimeout(context.Background(), bsc.Timeout)
	defer cancel()

	nodes, err := bsc.Client.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("request to list nodes failed: %w", err)
	}

	out, err := format.MarshalData(nodes, outFormat)
	if err != nil {
		return nil, fmt.Errorf("formatting node list failed: %w", err)
	}

	return out, nil
}
