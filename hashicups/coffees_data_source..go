package hashicups

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &coffeesDataSource{}
	_ datasource.DataSourceWithConfigure = &coffeesDataSource{}
)

func NewCoffeesDataSource() datasource.DataSource {
	return &coffeesDataSource{}
}

type coffeesDataSource struct {
	client *Client
}

// coffeesDataSourceModel maps the data source schema data.
type coffeesDataSourceModel struct {
	ID      types.String   `tfsdk:"id"`
	Coffees []coffeesModel `tfsdk:"coffees"`
}

// coffeesModel maps coffees schema data.
type coffeesModel struct {
	ID          types.Int64               `tfsdk:"id"`
	Name        types.String              `tfsdk:"name"`
	Teaser      types.String              `tfsdk:"teaser"`
	Description types.String              `tfsdk:"description"`
	Price       types.Float64             `tfsdk:"price"`
	Image       types.String              `tfsdk:"image"`
	Ingredients []coffeesIngredientsModel `tfsdk:"ingredients"`
}

// coffeesIngredientsModel maps coffee ingredients data
type coffeesIngredientsModel struct {
	ID types.Int64 `tfsdk:"id"`
}

func (c *coffeesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_coffees"
}

// Schema defines the schema for the data source.
func (c *coffeesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Fetches the list of coffees.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder identifier attribute.",
			},
			"coffees": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of coffees.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Numeric identifier of the coffee.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Product name of the coffee.",
							Computed:    true,
						},
						"teaser": schema.StringAttribute{
							Description: "Fun tagline for the coffee.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Product description of the coffee.",
							Computed:    true,
						},
						"price": schema.Float64Attribute{
							Description: "Suggested cost of the coffee.",
							Computed:    true,
						},
						"image": schema.StringAttribute{
							Description: "URI for an image of the coffee.",
							Computed:    true,
						},
						"ingredients": schema.ListNestedAttribute{
							Description: "List of ingredients in the coffee.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.Int64Attribute{
										Description: "Numeric identifier of the coffee ingredient.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (c *coffeesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state coffeesDataSourceModel

	coffees, err := c.client.GetCoffees()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read HashiCups Coffees",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, coffee := range coffees {
		coffeeState := coffeesModel{
			ID:          types.Int64Value(int64(coffee.ID)),
			Name:        types.StringValue(coffee.Name),
			Teaser:      types.StringValue(coffee.Teaser),
			Description: types.StringValue(coffee.Description),
			Price:       types.Float64Value(coffee.Price),
			Image:       types.StringValue(coffee.Image),
		}

		for _, ingredient := range coffee.Ingredient {
			coffeeState.Ingredients = append(coffeeState.Ingredients, coffeesIngredientsModel{
				ID: types.Int64Value(int64(ingredient.ID)),
			})
		}

		state.Coffees = append(state.Coffees, coffeeState)
	}

	state.ID = types.StringValue("placeholder")

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (c *coffeesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	c.client = request.ProviderData.(*Client)
}
