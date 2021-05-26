package main

import (
	"context"
	"fmt"
	"os"

	"batchtask/config"

	"github.com/Azure/azure-sdk-for-go/services/batch/2017-05-01.5.0/batch"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	tenantId     = ""
	clientId     = ""
	clientSecret = ""
)

func main() {

	// for auth

	oauthConfig, err := adal.NewOAuthConfig(
		config.Environment().ActiveDirectoryEndpoint, tenantId)
	if err != nil {
		fmt.Printf("init auth fails")
		os.Exit(1)
	}

	token, err := adal.NewServicePrincipalToken(
		*oauthConfig, clientId, clientSecret, config.Environment().BatchManagementEndpoint)
	if err != nil {
		fmt.Printf("init auth token fails: %v", err)
		os.Exit(1)
	}

	a := autorest.NewBearerAuthorizer(token)

	//=====
	ctx := context.Background()

	taskID := "myprog1"
	taskClient := batch.NewTaskClientWithBaseURI("https://mydev.japanwest.batch.azure.com")

	taskClient.Authorizer = a

	taskToAdd := batch.TaskAddParameter{
		ID:          &taskID,
		CommandLine: to.StringPtr("/bin/bash -c 'set -e; set -o pipefail; echo Hello world from the Batch Hello world sample!; wait'"),
		UserIdentity: &batch.UserIdentity{
			AutoUser: &batch.AutoUserSpecification{
				ElevationLevel: batch.Admin,
				Scope:          batch.Task,
			},
		},
	}

	_, err = taskClient.Add(ctx, "zap", taskToAdd, nil, nil, nil, nil)

	if err != nil {
		fmt.Printf("error: %v", err)
	}
}
