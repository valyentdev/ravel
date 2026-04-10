package endpoints

import (
	"context"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/internal/id"
)

type SecretResolver struct {
	NSResolver
	SecretName string `path:"secret_name"`
}

type CreateSecretRequest struct {
	NSResolver
	Body *api.CreateSecretPayload
}

type CreateSecretResponse struct {
	Body *api.Secret
}

func (e *Endpoints) createSecret(ctx context.Context, req *CreateSecretRequest) (*CreateSecretResponse, error) {
	secretId := id.Generate()
	err := e.ravel.State.Queries.CreateSecret(ctx, secretId, req.Body.Name, req.Namespace, req.Body.Value)
	if err != nil {
		e.log("Failed to create secret", err)
		return nil, err
	}

	secret, err := e.ravel.State.Queries.GetSecret(ctx, req.Namespace, req.Body.Name)
	if err != nil {
		e.log("Failed to get created secret", err)
		return nil, err
	}

	return &CreateSecretResponse{
		Body: &secret,
	}, nil
}

type ListSecretsRequest struct {
	NSResolver
}

type ListSecretsResponse struct {
	Body []api.Secret `json:"secrets"`
}

func (e *Endpoints) listSecrets(ctx context.Context, req *ListSecretsRequest) (*ListSecretsResponse, error) {
	secrets, err := e.ravel.State.Queries.ListSecrets(ctx, req.Namespace)
	if err != nil {
		e.log("Failed to list secrets", err)
		return nil, err
	}

	if secrets == nil {
		secrets = []api.Secret{}
	}

	return &ListSecretsResponse{
		Body: secrets,
	}, nil
}

type GetSecretRequest struct {
	SecretResolver
}

type GetSecretResponse struct {
	Body *api.Secret
}

func (e *Endpoints) getSecret(ctx context.Context, req *GetSecretRequest) (*GetSecretResponse, error) {
	secret, err := e.ravel.State.Queries.GetSecret(ctx, req.Namespace, req.SecretName)
	if err != nil {
		e.log("Failed to get secret", err)
		return nil, err
	}

	return &GetSecretResponse{
		Body: &secret,
	}, nil
}

type UpdateSecretRequest struct {
	SecretResolver
	Body *api.UpdateSecretPayload
}

type UpdateSecretResponse struct {
	Body *api.Secret
}

func (e *Endpoints) updateSecret(ctx context.Context, req *UpdateSecretRequest) (*UpdateSecretResponse, error) {
	err := e.ravel.State.Queries.UpdateSecret(ctx, req.Namespace, req.SecretName, req.Body.Value)
	if err != nil {
		e.log("Failed to update secret", err)
		return nil, err
	}

	secret, err := e.ravel.State.Queries.GetSecret(ctx, req.Namespace, req.SecretName)
	if err != nil {
		e.log("Failed to get updated secret", err)
		return nil, err
	}

	return &UpdateSecretResponse{
		Body: &secret,
	}, nil
}

type DeleteSecretRequest struct {
	SecretResolver
}

type DeleteSecretResponse struct {
}

func (e *Endpoints) deleteSecret(ctx context.Context, req *DeleteSecretRequest) (*DeleteSecretResponse, error) {
	err := e.ravel.State.Queries.DeleteSecret(ctx, req.Namespace, req.SecretName)
	if err != nil {
		e.log("Failed to delete secret", err)
		return nil, err
	}

	return nil, nil
}
