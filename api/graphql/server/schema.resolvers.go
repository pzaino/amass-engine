package server

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.40

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/owasp-amass/config/config"
	"github.com/owasp-amass/engine/api/graphql/server/model"
	"github.com/owasp-amass/engine/sessions"
	"github.com/owasp-amass/engine/types"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context, input model.CreateSessionInput) (*model.Session, error) {
	fmt.Println("Create Session Called")

	//r.scheduler.Schedule()
	//input.Config

	testSession := &model.Session{
		SessionToken: "00000000-0000-0000-0000-0000000000033", //?
	}
	return testSession, nil
}

// CreateSessionFromJSON is the resolver for the createSessionFromJson field.
func (r *mutationResolver) CreateSessionFromJSON(ctx context.Context, input model.CreateSessionJSONInput) (*model.Session, error) {
	fmt.Println("CreateSessionFromJSON")

	var config config.Config
	err := json.Unmarshal([]byte(input.Config), &config)
	if err != nil {
		fmt.Println(err)
	}

	// Populate FROM/TO in transformations
	for k, t := range config.Transformations {
		t.Split(k)
	}

	newSession, err := sessions.NewSession(&config)
	if err != nil {
		fmt.Println(err)
	}

	token, err := r.sessionManager.Add(newSession)
	if err != nil {
		fmt.Println(err)
	}

	model := &model.Session{
		SessionToken: token.String(),
	}
	return model, nil
}

// CreateAsset is the resolver for the createAsset field.
func (r *mutationResolver) CreateAsset(ctx context.Context, input model.CreateAssetInput) (*model.Asset, error) {
	token, _ := uuid.Parse(input.SessionToken)
	session := r.sessionManager.Get(token)
	if session == nil {
		return nil, errors.New("invalid session")
	}

	data, err := json.Marshal(input.Data)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("invalid json data")
	}

	// Unmarshal json into AssetData struct
	var assetData types.AssetData
	err = json.Unmarshal(data, &assetData)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("invalid json data")
	}

	// Create and schedule new event
	event := &types.Event{
		UUID:      uuid.New(),
		Name:      *input.AssetName,
		SessionID: token,
		Data:      assetData,
		Type:      types.EventTypeAsset,
	}
	err = r.sched.Schedule(event)
	if err != nil {
		return nil, errors.New("failed to create asset")
	}

	// TEST: Broadcast asset creation message to client subscribers
	session.PubSub.Publish("Log created asset")

	model := &model.Asset{
		ID: event.UUID.String(),
	}
	return model, nil
}

// TerminateSession is the resolver for the terminateSession field.
func (r *mutationResolver) TerminateSession(ctx context.Context, sessionToken string) (*bool, error) {
	result := false
	token, _ := uuid.Parse(sessionToken)

	if r.sessionManager.Get(token) != nil {
		r.sessionManager.Cancel(token)
		result = true
	} else {
		return &result, errors.New("invalid session token")
	}

	return &result, nil
}

// SessionStats is the resolver for the sessionStats field.
func (r *queryResolver) SessionStats(ctx context.Context, sessionToken string) (*model.SessionStats, error) {
	token, _ := uuid.Parse(sessionToken)

	if r.sessionManager.Get(token) == nil {
		return nil, errors.New("invalid session token")
	}

	ssResp := r.sched.GetSessionStats(token, types.EventTypeAsset)

	model := &model.SessionStats{
		WorkItemsInProcess:   &ssResp.SessionWorkItemsInProcess,
		WorkItemsWaiting:     &ssResp.SessionWorkItemsWaiting,
		WorkItemsProcessable: &ssResp.SessionWorkItemsProcessable,
	}

	return model, nil
}

// SystemStats is the resolver for the systemStats field.
func (r *queryResolver) SystemStats(ctx context.Context, sessionToken string) (*model.SystemStats, error) {
	panic(fmt.Errorf("not implemented: SystemStats - systemStats"))
}

// LogMessages is the resolver for the logMessages field.
func (r *subscriptionResolver) LogMessages(ctx context.Context, sessionToken string) (<-chan *string, error) {
	token, _ := uuid.Parse(sessionToken)
	session := r.sessionManager.Get(token)

	fmt.Println("LogMessages callled")
	if session != nil {

		session.PubSub.Publish("Channel created")
		ch := session.PubSub.Subscribe()

		return ch, nil
	}

	return nil, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
