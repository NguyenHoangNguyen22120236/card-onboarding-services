package store

import (
	"context"
	"errors"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/observability"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamodbAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

type DynamoDBRequestStatusStore struct {
	client    dynamodbAPI
	tableName string
}

func NewDynamoDBRequestStatusStore(client dynamodbAPI, tableName string) *DynamoDBRequestStatusStore {
	return &DynamoDBRequestStatusStore{
		client:    client,
		tableName: tableName,
	}
}

func (s *DynamoDBRequestStatusStore) GetByCustomerID(ctx context.Context, customerID string) (entity.RequestStatus, error) {
	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key:       customerIDKey(customerID),
	})
	if err != nil {
		return entity.RequestStatus{}, err
	}
	if len(out.Item) == 0 {
		return entity.RequestStatus{}, ErrNotFound
	}

	return unmarshalRequestStatusItem(out.Item)
}

func (s *DynamoDBRequestStatusStore) Save(ctx context.Context, status entity.RequestStatus) error {
	item, err := marshalRequestStatusItem(status)
	if err != nil {
		logDBWriteFailed("request_status_save", s.tableName, status.CustomerID, err)
		return err
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	if err != nil {
		logDBWriteFailed("request_status_save", s.tableName, status.CustomerID, err)
	}
	return err
}

func (s *DynamoDBRequestStatusStore) Update(ctx context.Context, status entity.RequestStatus) error {
	item, err := marshalRequestStatusItem(status)
	if err != nil {
		logDBWriteFailed("request_status_update", s.tableName, status.CustomerID, err)
		return err
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(customerId)"),
	})
	if isConditionalCheckFailed(err) {
		logDBWriteFailed("request_status_update", s.tableName, status.CustomerID, err)
		return ErrNotFound
	}
	if err != nil {
		logDBWriteFailed("request_status_update", s.tableName, status.CustomerID, err)
	}
	return err
}

func marshalRequestStatusItem(status entity.RequestStatus) (map[string]types.AttributeValue, error) {
	return attributevalue.MarshalMap(status)
}

func unmarshalRequestStatusItem(item map[string]types.AttributeValue) (entity.RequestStatus, error) {
	var status entity.RequestStatus
	err := attributevalue.UnmarshalMap(item, &status)
	return status, err
}

func customerIDKey(customerID string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"customerId": &types.AttributeValueMemberS{Value: customerID},
	}
}

func isConditionalCheckFailed(err error) bool {
	var conditionalErr *types.ConditionalCheckFailedException
	return errors.As(err, &conditionalErr)
}

func logDBWriteFailed(operation string, tableName string, customerID string, err error) {
	fields := observability.NewFields()
	fields.Operation = operation
	fields.TableName = tableName
	fields.CustomerID = customerID
	if err != nil {
		fields.ErrorMessage = err.Error()
	}

	observability.LogCount(observability.MetricDBWriteFailedCount, fields)
}
