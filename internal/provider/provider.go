// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	/*
		Endpoint types.String `tfsdk:"endpoint"`
	*/
}

func (p *AccelByteProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "accelbyte"
	resp.Version = p.version
}

func (p *AccelByteProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		/*
			Attributes: map[string]schema.Attribute{
				"endpoint": schema.StringAttribute{
					MarkdownDescription: "Example provider attribute",
					Optional:            true,
				},
			},
		*/
	}
}

func (p *AccelByteProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AccelByteProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
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
