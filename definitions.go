package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
)

func getPolicyClient(subscriptionID string) (*armpolicy.DefinitionsClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := armpolicy.NewDefinitionsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func CreateOrUpdatePolicyDefinition(ctx context.Context, client *armpolicy.DefinitionsClient, p Policy) (string, error) {
	if p.Definition.Properties == nil {
		return "", errors.New("policy definition properties is required")
	}
	
	if p.Definition.Properties.DisplayName == nil {
		return "", errors.New("policy definition display name is required")
	}
	managementGroupID := ""
	if p.Definition.Properties.Metadata != nil {
		if meta, ok := p.Definition.Properties.Metadata.(map[string]any); ok {
			if val, ok := meta["managementGroupId"]; ok && val != nil && val != "" {
				managementGroupID = val.(string)
			}
		}
	}

	definition := armpolicy.Definition{
		Properties: p.Definition.Properties,
	}
	var definitionID string


	if managementGroupID != "" {
		r, err := client.CreateOrUpdateAtManagementGroup(
			ctx,
			managementGroupID,
			*p.Definition.Properties.DisplayName,
			definition,
			nil,
		)
		if err != nil {
			fmt.Println("CreateOrUpdateAtManagementGroup failed:", err)
			return "", err
		}
		out, err := r.MarshalJSON()
		if err == nil {
			fmt.Println("Created Policy Definition (Management Group):", string(out))
		}
		if r.ID != nil {
			definitionID = *r.ID
		}
	} else {
		r, err := client.CreateOrUpdate(
			ctx,
			*p.Definition.Properties.DisplayName,
			definition,
			nil,
		)
		if err != nil {
			fmt.Println("CreateOrUpdate failed:", err)
			return "", err
		}
		out, err := r.MarshalJSON()
		if err == nil {
			fmt.Println("Created Policy Definition:", string(out))
		}
		if r.ID != nil {
			definitionID = *r.ID
		}
	}

	if definitionID == "" {
		return "", errors.New("policy definition not created")
	}

	return definitionID, nil
}

func DeletePolicyDefinition(ctx context.Context, client *armpolicy.DefinitionsClient, p Policy) error {
	if p.Definition.Properties == nil {
		return errors.New("policy definition properties is required")
	}

	if p.Definition.Properties.DisplayName == nil {
		return errors.New("policy definition display name is required")
	}

	managementGroupID := ""
	if p.Definition.Properties.Metadata != nil {
		if meta, ok := p.Definition.Properties.Metadata.(map[string]any); ok {
			if val, ok := meta["managementGroupId"]; ok && val != nil && val != "" {
				managementGroupID = val.(string)
			}
		}
	}

	if managementGroupID != "" {
		_, err := client.DeleteAtManagementGroup(ctx, managementGroupID, *p.Definition.Properties.DisplayName, nil)
		return err
	}
	_, err := client.Delete(ctx, *p.Definition.Properties.DisplayName, nil)
	return err
}
