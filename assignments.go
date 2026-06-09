package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
)

func getAssignmentsClient(subscriptionID string) (*armpolicy.AssignmentsClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := armpolicy.NewAssignmentsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// CreateOrUpdatePolicyAssignment creates or updates a policy assignment
func CreateOrUpdatePolicyAssignment(ctx context.Context, client *armpolicy.AssignmentsClient, p Policy, policyDefinitionID string) (string, error) {
	if p.Assignment.Properties == nil {
		return "", errors.New("policy assignment properties is required")
	}

	if p.Assignment.Properties.DisplayName == nil {
		return "", errors.New("policy assignment display name is required")
	}

	if p.Assignment.Properties.Scope == nil {
		return "", errors.New("policy assignment scope is required")
	}

	assignment := armpolicy.Assignment{
		Properties: p.Assignment.Properties,
	}
	assignment.Properties.PolicyDefinitionID = &policyDefinitionID

	if p.Assignment.Identity != nil {
		assignment.Identity = p.Assignment.Identity
	}

	if p.Assignment.Location != nil {
		assignment.Location = p.Assignment.Location
	}

	resp, err := client.Create(ctx, *p.Assignment.Properties.Scope, *p.Assignment.Properties.DisplayName, assignment, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create policy assignment: %w", err)
	}

	out, err := resp.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy assignment response: %w", err)
	}

	return string(out), nil
}

// DeletePolicyAssignment deletes a policy assignment
func DeletePolicyAssignment(ctx context.Context, client *armpolicy.AssignmentsClient, policy Policy) error {
	if policy.Assignment.Properties == nil {
		return errors.New("policy assignment properties is required")
	}

	if policy.Assignment.Properties.DisplayName == nil {
		return errors.New("policy assignment display name is required")
	}

	if policy.Assignment.Properties.Scope == nil {
		return errors.New("policy assignment scope is required")
	}

	_, err := client.Delete(ctx, *policy.Assignment.Properties.Scope, *policy.Assignment.Properties.DisplayName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete policy assignment: %w", err)
	}
	return nil
}
