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

func runTask(taskID, targetUrl, targetDomain string) {
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

	taskClient := batch.NewTaskClientWithBaseURI("https://mydev.japanwest.batch.azure.com")

	taskClient.Authorizer = a

	// env
	envSetting := []batch.EnvironmentSetting{}

	keyTarget := "target"

	keyDomain := "domain"
	targetDomain = "github.com"

	keyReportFile := "reportfile"
	reportfile := taskID + "_report"

	envSetting = append(envSetting, batch.EnvironmentSetting{Name: &keyTarget, Value: &targetUrl})
	envSetting = append(envSetting, batch.EnvironmentSetting{Name: &keyDomain, Value: &targetDomain})
	envSetting = append(envSetting, batch.EnvironmentSetting{Name: &keyReportFile, Value: &reportfile})

	username := "zap"
	taskToAdd := batch.TaskAddParameter{
		ID:          &taskID,
		CommandLine: to.StringPtr("/bin/bash -c '
		mkdir -p report && 
		docker login -u ccf5e28e-a89e-4538-96f8-f08b206602b7 -p 4Tr.fL-xBXWOHKHy~_aE0x61c_n2x2ZH0- g7docker.azurecr.io && 
		docker run -d --rm --name zapapi g7docker.azurecr.io/owasp/zap2docker-stable:2.10.0 zap.sh -daemon -port 8080 -host 0.0.0.0 -config api.disablekey=true -config api.addrs.addr.name=.\\* -config api.addrs.addr.regex=true && sleep 15 && 
		docker run -i --rm --link zapapi:zapapi --name zaptool -v $PWD/report:/report -w /report g7docker.azurecr.io/cyber/zaptool:0.0.1 /root/zap -m 3 -M 1 -t $target -J $id -r $domain && 
		docker rm -f $(docker ps -a -q)' &&
		回報 java"),
		UserIdentity: &batch.UserIdentity{
			UserName: &username,
		},
		EnvironmentSettings: &envSetting,
	}

	_, err = taskClient.Add(ctx, "zap", taskToAdd, nil, nil, nil, nil)

	if err != nil {
		fmt.Printf("error: %v", err)
	}
}
func main() {

	targetUrl := "https://github.com/genchilu/algorithmPractice/blob/master/%20RangeSumQuery2DImmutable/golang/%20RangeSumQuery2D.go"
	domain := "github.com"

	//single task
	//runTask("demo", targetUrl, domain)

	// multi task
	taskIDPrefix := "poc"
	for i := 2; i < 20; i++ {
		taskID := fmt.Sprintf("%s_%d", taskIDPrefix, i)
		fmt.Printf("Submit task: %s\n", taskID)

		runTask(taskID, targetUrl, domain)
	}

}
