package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/devopsarr/terraform-provider-sonarr/tools"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golift.io/starr/sonarr"
)

const releaseProfilesDataSourceName = "release_profiles"

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ReleaseProfilesDataSource{}

func NewReleaseProfilesDataSource() datasource.DataSource {
	return &ReleaseProfilesDataSource{}
}

// ReleaseProfilesDataSource defines the release profiles implementation.
type ReleaseProfilesDataSource struct {
	client *sonarr.Sonarr
}

// ReleaseProfiles describes the release profiles data model.
type ReleaseProfiles struct {
	ReleaseProfiles types.Set    `tfsdk:"release_profiles"`
	ID              types.String `tfsdk:"id"`
}

func (d *ReleaseProfilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + releaseProfilesDataSourceName
}

func (d *ReleaseProfilesDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the release server.
		MarkdownDescription: "<!-- subcategory:Profiles -->List all available [Release Profiles](../resources/release_profile).",
		Attributes: map[string]tfsdk.Attribute{
			// TODO: remove ID once framework support tests without ID https://www.terraform.io/plugin/framework/acctests#implement-id-attribute
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"release_profiles": {
				MarkdownDescription: "Release Profile list.",
				Computed:            true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						MarkdownDescription: "Release Profile ID.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"enabled": {
						MarkdownDescription: "Enabled",
						Computed:            true,
						Type:                types.BoolType,
					},
					"name": {
						MarkdownDescription: "Release profile name.",
						Computed:            true,
						Type:                types.StringType,
					},
					"indexer_id": {
						MarkdownDescription: "Indexer ID. Set `0` for all.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"required": {
						MarkdownDescription: "Required terms.",
						Computed:            true,
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
					"ignored": {
						MarkdownDescription: "Ignored terms.",
						Computed:            true,
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
					"tags": {
						MarkdownDescription: "List of associated tags.",
						Computed:            true,
						Type: types.SetType{
							ElemType: types.Int64Type,
						},
					},
				}),
			},
		},
	}, nil
}

func (d *ReleaseProfilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sonarr.Sonarr)
	if !ok {
		resp.Diagnostics.AddError(
			tools.UnexpectedDataSourceConfigureType,
			fmt.Sprintf("Expected *sonarr.Sonarr, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ReleaseProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *ReleaseProfiles

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Get releaseprofiles current value
	response, err := d.client.GetReleaseProfilesContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(tools.ClientError, fmt.Sprintf("Unable to read %s, got error: %s", releaseProfileResourceName, err))

		return
	}

	tflog.Trace(ctx, "read "+releaseProfileResourceName)
	// Map response body to resource schema attribute
	profiles := make([]ReleaseProfile, len(response))
	for i, p := range response {
		profiles[i].write(ctx, p)
	}

	tfsdk.ValueFrom(ctx, profiles, data.ReleaseProfiles.Type(context.Background()), &data.ReleaseProfiles)
	// TODO: remove ID once framework support tests without ID https://www.terraform.io/plugin/framework/acctests#implement-id-attribute
	data.ID = types.StringValue(strconv.Itoa(len(response)))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
