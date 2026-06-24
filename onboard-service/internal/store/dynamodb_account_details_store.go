package store

import (
	"context"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBAccountDetailsStore struct {
	client    dynamodbAPI
	tableName string
}

func NewDynamoDBAccountDetailsStore(client dynamodbAPI, tableName string) *DynamoDBAccountDetailsStore {
	return &DynamoDBAccountDetailsStore{
		client:    client,
		tableName: tableName,
	}
}

func (s *DynamoDBAccountDetailsStore) GetByCustomerID(ctx context.Context, customerID string) (entity.AccountDetails, error) {
	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key:       customerIDKey(customerID),
	})
	if err != nil {
		return entity.AccountDetails{}, err
	}
	if len(out.Item) == 0 {
		return entity.AccountDetails{}, ErrNotFound
	}

	return unmarshalAccountDetailsItem(out.Item)
}

func (s *DynamoDBAccountDetailsStore) Save(ctx context.Context, details entity.AccountDetails) error {
	item, err := marshalAccountDetailsItem(details)
	if err != nil {
		logDBWriteFailed("account_details_save", s.tableName, details.CustomerID, err)
		return err
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	if err != nil {
		logDBWriteFailed("account_details_save", s.tableName, details.CustomerID, err)
	}
	return err
}

func (s *DynamoDBAccountDetailsStore) Update(ctx context.Context, details entity.AccountDetails) error {
	item, err := marshalAccountDetailsItem(details)
	if err != nil {
		logDBWriteFailed("account_details_update", s.tableName, details.CustomerID, err)
		return err
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(customerId)"),
	})
	if isConditionalCheckFailed(err) {
		logDBWriteFailed("account_details_update", s.tableName, details.CustomerID, err)
		return ErrNotFound
	}
	if err != nil {
		logDBWriteFailed("account_details_update", s.tableName, details.CustomerID, err)
	}
	return err
}

func marshalAccountDetailsItem(details entity.AccountDetails) (map[string]types.AttributeValue, error) {
	return attributevalue.MarshalMap(details)
}

func unmarshalAccountDetailsItem(item map[string]types.AttributeValue) (entity.AccountDetails, error) {
	var details entity.AccountDetails
	err := attributevalue.UnmarshalMap(item, &details)
	return details, err
}
