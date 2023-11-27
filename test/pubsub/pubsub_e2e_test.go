/*
Copyright 2022 The Numaproj Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pubsub

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/numaproj-contrib/numaflow-utils-go/testing/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	SUBSCRIPTION_ID = "subscription-09098ui1"
	PROJECT_ID      = "pubsub-test"
	TOPIC_ID        = "pubsub-test-topic"
	PUB_SUB_PORT    = 8681
)

type GCPPubSubSinkSuite struct {
	fixtures.E2ESuite
}

func isPubSubContainsMessages(ctx context.Context, client *pubsub.Client, msg string) bool {
	sub := client.Subscription(SUBSCRIPTION_ID)
	cancelContext, cancel := context.WithCancel(ctx)
	defer cancel()
	sendMessage := make(chan struct{}, 1)
	msgCount := 0
	go func() {
		err := sub.Receive(cancelContext, func(ctx context.Context, message *pubsub.Message) {
			if msgCount > 0 {
				sendMessage <- struct{}{}
			}
			msgCount++
		})
		if err != nil {
			log.Fatalf("error receiving messages %s", err)
		}
	}()
	select {
	case <-sendMessage:
		log.Println("returning true")

		return true
	case <-ctx.Done():
		log.Println("Context has bee n done")
		return false
	}
}
func ensureTopicAndSubscription(ctx context.Context, client *pubsub.Client, topicID, subID string) error {
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		if _, err = client.CreateTopic(ctx, topicID); err != nil {
			return err
		}
	}
	sub := client.Subscription(subID)
	exists, err = sub.Exists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		if _, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: topic}); err != nil {
			return err
		}
	}
	return nil
}

func createPubSubClient() *pubsub.Client {
	pubsubClient, err := pubsub.NewClient(context.Background(), PROJECT_ID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return pubsubClient
}
func (suite *GCPPubSubSinkSuite) SetupTest() {
	gcloudPubSubDeleteCmd := fmt.Sprintf("kubectl delete -k ../../config/apps/gcloud-pubsub -n %s --ignore-not-found=true", fixtures.Namespace)
	suite.Given().When().Exec("sh", []string{"-c", gcloudPubSubDeleteCmd}, fixtures.OutputRegexp(""))
	gcloudPubSubCreateCmd := fmt.Sprintf("kubectl apply -k ../../config/apps/gcloud-pubsub -n %s", fixtures.Namespace)
	suite.Given().When().Exec("sh", []string{"-c", gcloudPubSubCreateCmd}, fixtures.OutputRegexp("service/gcloud-pubsub created"))
	gcloudPubSubLabelSelector := fmt.Sprintf("app=%s", "gcloud-pubsub")
	suite.Given().When().WaitForStatefulSetReady(gcloudPubSubLabelSelector)
	suite.T().Log("gcloud-pubsub resources are ready")
	//delay to make system ready in CI
	time.Sleep(2 * time.Minute)
	suite.T().Log("port forwarding gcloud-pubsub service")
	suite.StartPortForward("gcloud-pubsub-0", PUB_SUB_PORT)
}

func (suite *GCPPubSubSinkSuite) TestPubSubSource() {
	err := os.Setenv("PUBSUB_EMULATOR_HOST", fmt.Sprintf("localhost:%d", PUB_SUB_PORT))
	var message = "testing"
	assert.Nil(suite.T(), err)
	pubSubClient := createPubSubClient()
	assert.NotNil(suite.T(), pubSubClient)
	err = ensureTopicAndSubscription(context.Background(), pubSubClient, TOPIC_ID, SUBSCRIPTION_ID)
	assert.Nil(suite.T(), err)
	workflow := suite.Given().Pipeline("@testdata/pubsub_sink.yaml").
		When().
		CreatePipelineAndWait()
	workflow.Expect().VertexPodsRunning()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	containMsg := isPubSubContainsMessages(ctx, pubSubClient, message)
	suite.True(containMsg)
	workflow.DeletePipelineAndWait(3 * time.Minute)

}

func TestGCPPubSubSourceSuite(t *testing.T) {
	suite.Run(t, new(GCPPubSubSinkSuite))
}
