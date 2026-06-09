package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
)

var data = `
{
  "definition": {
    "properties": {
	  "displayName": "sample-audit-policy-three",
      "description": "testing",
	  "policyType": "Custom",
      "mode": "All",
      "metadata": {
	  	"policyId": "123",
        "category": "test",
        "tenant-id": "13",
		"managementGroupId": ""
      },
      "policyRule": {
        "if": {
          "allOf": [
            {
              "field": "type",
              "equals": "Microsoft.Storage/storageAccounts"
            },
            {
              "field": "Microsoft.Storage/storageAccounts/allowBlobPublicAccess",
              "equals": true
            }
          ]
        },
        "then": {
          "effect": "audit"
        }
      }
    }
  },
  "assignment": {
    "properties": {
      "description": "test description for assignment",
      "displayName": "sample-audit-policy-three-assignment",
      "enforcementMode": "DoNotEnforce",
      "scope": "/subscriptions/8daf11df-39be-4fc4-af77-7b2ae0d3866e",
	   "metadata": {
	    "policyId": "123",
        "tenant-id": "13"
      }
    }
  },
  "action": "DELETE"
}`

type Policy struct {
	Action     string `json:"action"` // create | update | delete
	Definition armpolicy.Definition
	Assignment armpolicy.Assignment
}

func policyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var policyDetails Policy
	err := json.NewDecoder(r.Body).Decode(&policyDetails)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		fmt.Fprintf(os.Stderr, "AZURE_SUBSCRIPTION_ID is not set")
		http.Error(w, "AZURE_SUBSCRIPTION_ID is not set", http.StatusInternalServerError)
		return
	}

	defClient, err := getPolicyClient(subscriptionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get policy client: %v", err), http.StatusInternalServerError)
		return
	}

	assignClient, err := getAssignmentsClient(subscriptionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get assignments client: %v", err), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	actionUpper := strings.ToUpper(policyDetails.Action)

	type Response struct {
		Message      string `json:"message"`
		DefinitionID string `json:"definitionId,omitempty"`
	}

	switch actionUpper {
	case "CREATE", "UPDATE":
		// 1. Create/Update the policy definition
		defID, err := CreateOrUpdatePolicyDefinition(ctx, defClient, policyDetails)
		if err != nil {
			http.Error(w, fmt.Sprintf("Create/Update Definition failed: %v", err), http.StatusInternalServerError)
			return
		}

		// 2. Assign the policy definition
		_, err = CreateOrUpdatePolicyAssignment(ctx, assignClient, policyDetails, defID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Create/Update Assignment failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Message:      "Policy definition and assignment created/updated successfully",
			DefinitionID: defID,
		})

	case "DELETE":
		// 1. Delete the policy assignment first
		err = DeletePolicyAssignment(ctx, assignClient, policyDetails)
		if err != nil {
			log.Println("Delete Assignment failed (it might not exist): ", err)
		}

		// 2. Delete the policy definition
		err = DeletePolicyDefinition(ctx, defClient, policyDetails)
		if err != nil {
			http.Error(w, fmt.Sprintf("Delete Definition failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Message: "Policy definition and assignment deleted successfully",
		})

	default:
		http.Error(w, fmt.Sprintf("Invalid action: %s", policyDetails.Action), http.StatusBadRequest)
	}
}

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}

	http.HandleFunc("/api/policy", policyHandler)

	log.Printf("About to listen on %s. Go to http://127.0.0.1%s/", listenAddr, listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
