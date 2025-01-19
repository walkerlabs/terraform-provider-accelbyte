// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure AccelByteProvider satisfies various provider interfaces.
var _ provider.Provider = &AccelByteProvider{}
var _ provider.ProviderWithFunctions = &AccelByteProvider{}
var _ provider.ProviderWithEphemeralResources = &AccelByteProvider{}

// AccelByteProvider defines the provider implementation.
type AccelByteProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AccelByteProviderModel describes the provider data model.
type AccelByteProviderModel struct {
	BaseUrl         types.String `tfsdk:"base_url"`
	IamClientId     types.String `tfsdk:"iam_client_id"`
	IamClientSecret types.String `tfsdk:"iam_client_secret"`
	AdminUsername   types.String `tfsdk:"admin_username"`
	AdminPassword   types.String `tfsdk:"admin_password"`
}

func (p *AccelByteProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "accelbyte"
	resp.Version = p.version
}

func (p *AccelByteProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "URL to AccelByte cluster, typically on the form 'https://<something>.accelbyte.io'",
				Required:            true,
			},
			"iam_client_id": schema.StringAttribute{
				MarkdownDescription: "IAM Client ID to use for authentication",
				Required:            true,
			},
			"iam_client_secret": schema.StringAttribute{
				MarkdownDescription: "IAM Client Secret to use for authentication",
				Required:            true,
				Sensitive:           true,
			},
			"admin_username": schema.StringAttribute{
				MarkdownDescription: "Admin user email to use for authentication",
				Required:            true,
			},
			"admin_password": schema.StringAttribute{
				MarkdownDescription: "Admin user password to use for authentication",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *AccelByteProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AccelByteProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if data.BaseUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown AccelByte API base_url",
			"The provider cannot create the AccelByte API client as the base_url nas not been given.. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ACCELBYTE_BASE_URL environment variable.",
		)
	}

	if data.IamClientId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_client_id"),
			"Unknown AccelByte API iam_client_id",
			"The provider cannot create the AccelByte API client as the iam_client_id nas not been given.. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ACCELBYTE_IAM_CLIENT_ID environment variable.",
		)
	}

	if data.IamClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_client_secret"),
			"Unknown AccelByte API iam_client_secret",
			"The provider cannot create the AccelByte API client as the iam_client_secret nas not been given.. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ACCELBYTE_IAM_CLIENT_SECRET environment variable.",
		)
	}

	if data.AdminUsername.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("admin_username"),
			"Unknown AccelByte API admin_username",
			"The provider cannot create the AccelByte API client as the admin_username nas not been given.. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ACCELBYTE_ADMIN_USERNAME environment variable.",
		)
	}

	if data.AdminPassword.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("admin_password"),
			"Unknown AccelByte API admin_password",
			"The provider cannot create the AccelByte API client as the admin_password nas not been given.. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ACCELBYTE_ADMIN_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	baseUrl := os.Getenv("ACCELBYTE_BASE_URL")
	iamClientId := os.Getenv("ACCELBYTE_IAM_CLIENT_ID")
	iamClientSecret := os.Getenv("ACCELBYTE_IAM_CLIENT_SECRET")
	adminUsername := os.Getenv("ACCELBYTE_ADMIN_USERNAME")
	adminPassword := os.Getenv("ACCELBYTE_ADMIN_PASSWORD")

	if !data.BaseUrl.IsNull() {
		baseUrl = data.BaseUrl.ValueString()
	}

	if !data.IamClientId.IsNull() {
		iamClientId = data.IamClientId.ValueString()
	}

	if !data.IamClientSecret.IsNull() {
		iamClientSecret = data.IamClientSecret.ValueString()
	}

	if !data.AdminUsername.IsNull() {
		adminUsername = data.AdminUsername.ValueString()
	}

	if !data.AdminPassword.IsNull() {
		adminPassword = data.AdminPassword.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if baseUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing HashiCups API base_url",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API base_url. "+
				"Set the base_url value in the configuration or use the ACCELBYTE_BASE_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if iamClientId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_client_id"),
			"Missing HashiCups API iam_client_id",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API iam_client_id. "+
				"Set the iam_client_id value in the configuration or use the ACCELBYTE_IAM_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if iamClientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_client_secret"),
			"Missing HashiCups API iam_client_secret",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API iam_client_secret. "+
				"Set the iam_client_secret value in the configuration or use the ACCELBYTE_IAM_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if adminUsername == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("admin_username"),
			"Missing HashiCups API admin_username",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API admin_username. "+
				"Set the admin_username value in the configuration or use the ACCELBYTE_ADMIN_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if adminPassword == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("admin_password"),
			"Missing HashiCups API admin_password",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API admin_password. "+
				"Set the admin_password value in the configuration or use the ACCELBYTE_ADMIN_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure Base URL (i.e. which cluster), and IAM Client/Secret pair
	// These will later on be used during the LoginUser call

	configRepository := auth.ConfigRepositoryImpl{
		ClientId:     iamClientId,
		ClientSecret: iamClientSecret,
		BaseUrl:      baseUrl,
	}

	tokenRepository := auth.DefaultTokenRepositoryImpl()

	oAuth20Service := &iam.OAuth20Service{
		Client:           factory.NewIamClient(&configRepository),
		ConfigRepository: &configRepository,
		TokenRepository:  tokenRepository,
		RefreshTokenRepository: &auth.RefreshTokenImpl{ // Automatically refresh the token when 50% of its lifetime has passed
			RefreshRate: 0.5,
			AutoRefresh: true,
		},
	}

	// Login to AccelByte backend, using admin username/password
	// This is the first backend API call, so this is the point where the following parameters are used for the first time (and thus get validated):
	// - Base URL
	// - IAM Client ID
	// - IAM Client Secret
	// - Admin Username
	// - Admin Password

	err := oAuth20Service.LoginUser(adminUsername, adminPassword)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to log in to AccelByte backend",
			"Login using admin username/password failed: "+
				err.Error(),
		)
		return
	}

	/*
		// Test SDK calls

		userProfileService := &basic.UserProfileService{
			Client:          factory.NewBasicClient(&configRepository),
			TokenRepository: tokenRepository,
		}
		input := &user_profile.GetUserProfileInfoParams{
			Namespace: "alaraprime",
			UserID:    "9765327f993e4e92b319bfa49f2ec956",
		}

		// Make a call to getMyProfileInfo endpoint through the wrapper
		userInfo, err := userProfileService.GetUserProfileInfoShort(input)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get profile info",
				"Unable to get profile info:"+err.Error(),
			)
			return
		}

		resp.Diagnostics.AddError(
			"got profile info!",
			"got profile info: "+userInfo.PublicID,
		)
	*/

	// resp.DataSourceData = client
	// resp.ResourceData = client
}

func (p *AccelByteProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *AccelByteProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewExampleEphemeralResource,
	}
}

func (p *AccelByteProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func (p *AccelByteProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AccelByteProvider{
			version: version,
		}
	}
}
