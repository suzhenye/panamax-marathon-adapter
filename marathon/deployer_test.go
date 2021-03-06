package marathon

import (
	"errors"
	"testing"
	"time"

	"github.com/CenturyLinkLabs/gomarathon"
	"github.com/CenturyLinkLabs/panamax-marathon-adapter/api"
	"github.com/stretchr/testify/assert"
)

var timeoutDuration = time.Second * 10

func testSuccessState(deployment *deployment, ctx *context) stateFn {
	deployment.status.code = OK
	return nil
}

func testFailState(deployment *deployment, ctx *context) stateFn {
	deployment.status.code = FAIL
	return nil
}

func testTimeoutState(deployment *deployment, ctx *context) stateFn {
	time.Sleep(time.Second * 15)
	deployment.status.code = OK
	return nil
}

func TestGroupDeployment(t *testing.T) {
	var deployment1, deployment2 deployment
	var deployer MarathonDeployer

	deployment1.name = "slowApp"
	deployment1.startingState = testSuccessState
	deployment2.name = "testApp"
	deployment2.startingState = testSuccessState

	myGroup := new(deploymentGroup)
	myGroup.deployments = []deployment{deployment1, deployment2}

	deployer.DeployGroup(myGroup, timeoutDuration)

	assert.Equal(t, true, myGroup.Done())
}

func TestFailedGroupDeployment(t *testing.T) {
	var deployment1, deployment2 deployment
	var deployer MarathonDeployer

	deployment1.name = "slowApp"
	deployment1.startingState = testSuccessState
	deployment2.name = "failApp"
	deployment2.startingState = testFailState

	myGroup := new(deploymentGroup)
	myGroup.deployments = []deployment{deployment1, deployment2}

	deployer.DeployGroup(myGroup, timeoutDuration)

	assert.Equal(t, true, myGroup.Failed())
}

func TestTimedoutGroupDeployment(t *testing.T) {
	var deployment1, deployment2 deployment
	var deployer MarathonDeployer

	deployment1.name = "testApp"
	deployment1.startingState = testSuccessState
	deployment2.name = "timeoutApp"
	deployment2.startingState = testTimeoutState

	myGroup := new(deploymentGroup)
	myGroup.deployments = []deployment{deployment1, deployment2}

	status := deployer.DeployGroup(myGroup, timeoutDuration)

	assert.Equal(t, OK, myGroup.deployments[0].status.code)
	assert.Equal(t, TIMEOUT, status.code)
}

func TestBuildDeploymentGroup(t *testing.T) {
	var deployer MarathonDeployer
	var testClient mockClient

	resp := new(gomarathon.Response)
	services := make([]*api.Service, 1)
	services[0] = &api.Service{Name: "foo"}

	deployer.uidGenerator = func() string { return "pmx" }

	testClient.On("GetGroup", "pmx").Return(resp, errors.New("test"))
	group := deployer.BuildDeploymentGroup(services, &testClient)

	assert.Equal(t, "/pmx/foo", group.deployments[0].application.ID)
}
