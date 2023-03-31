package hashicups

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &orderResource{}
	_ resource.ResourceWithConfigure   = &orderResource{}
	_ resource.ResourceWithImportState = &orderResource{}
)

type orderResource struct {
	client *Client
}

// orderResourceModel maps the resource schema data.
type orderResourceModel struct {
	ID          types.String     `tfsdk:"id"`
	Items       []orderItemModel `tfsdk:"items"`
	LastUpdated types.String     `tfsdk:"last_updated"`
}

// orderItemModel maps order item data.
type orderItemModel struct {
	Coffee   orderItemCoffeeModel `tfsdk:"coffee"`
	Quantity types.Int64          `tfsdk:"quantity"`
}

// orderItemCoffeeModel maps coffee order item data.
type orderItemCoffeeModel struct {
	ID          types.Int64   `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Teaser      types.String  `tfsdk:"teaser"`
	Description types.String  `tfsdk:"description"`
	Price       types.Float64 `tfsdk:"price"`
	Image       types.String  `tfsdk:"image"`
}

func NewOrderResource() resource.Resource {
	return &orderResource{}
}

func (o *orderResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_order"
}

// Schema defines the schema for the resource.
func (o *orderResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages an order.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Numeric identifier of the order.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the last Terraform update of the order.",
			},
			"items": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of items in the order.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"coffee": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Coffee item in the order.",
							Attributes: map[string]schema.Attribute{
								"id": schema.Int64Attribute{
									Description: "Numeric identifier of the coffee.",
									Required:    true,
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
							},
						},
						"quantity": schema.Int64Attribute{
							Required:    true,
							Description: "Count of this item in the order.",
						},
					},
				},
			},
		},
	}
}

func (o *orderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan orderResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var items []OrderItem
	for _, item := range plan.Items {
		items = append(items, OrderItem{
			Coffee: Coffee{
				ID: int(item.Coffee.ID.ValueInt64()),
			},
			Quantity: int(item.Quantity.ValueInt64()),
		})
	}

	order, err := o.client.CreateOrder(items)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating order",
			"An unexpected error was encountered trying to create the order."+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(order.ID))
	for orderItemIndex, orderItem := range order.Items {
		plan.Items[orderItemIndex] = orderItemModel{
			Coffee: orderItemCoffeeModel{
				ID:          types.Int64Value(int64(orderItem.Coffee.ID)),
				Name:        types.StringValue(orderItem.Coffee.Name),
				Teaser:      types.StringValue(orderItem.Coffee.Teaser),
				Description: types.StringValue(orderItem.Coffee.Description),
				Price:       types.Float64Value(orderItem.Coffee.Price),
				Image:       types.StringValue(orderItem.Coffee.Image),
			},
			Quantity: types.Int64Value(int64(orderItem.Quantity)),
		}
	}
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (o *orderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state orderResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	order, err := o.client.GetOrder(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading HashiCups Order",
			"Could not read HashiCups order ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Items = []orderItemModel{}
	for _, item := range order.Items {
		state.Items = append(state.Items, orderItemModel{
			Coffee: orderItemCoffeeModel{
				ID:          types.Int64Value(int64(item.Coffee.ID)),
				Name:        types.StringValue(item.Coffee.Name),
				Teaser:      types.StringValue(item.Coffee.Teaser),
				Description: types.StringValue(item.Coffee.Description),
				Price:       types.Float64Value(item.Coffee.Price),
				Image:       types.StringValue(item.Coffee.Image),
			},
			Quantity: types.Int64Value(int64(item.Quantity)),
		})
	}

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (o *orderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan orderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var hashicupsItems []OrderItem
	for _, item := range plan.Items {
		hashicupsItems = append(hashicupsItems, OrderItem{
			Coffee: Coffee{
				ID: int(item.Coffee.ID.ValueInt64()),
			},
			Quantity: int(item.Quantity.ValueInt64()),
		})
	}

	// Update existing order
	_, err := o.client.UpdateOrder(plan.ID.ValueString(), hashicupsItems)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating HashiCups Order",
			"Could not update order, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch updated items from GetOrder as UpdateOrder items are not
	// populated.
	order, err := o.client.GetOrder(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HashiCups Order",
			"Could not read HashiCups order ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.Items = []orderItemModel{}
	for _, item := range order.Items {
		plan.Items = append(plan.Items, orderItemModel{
			Coffee: orderItemCoffeeModel{
				ID:          types.Int64Value(int64(item.Coffee.ID)),
				Name:        types.StringValue(item.Coffee.Name),
				Teaser:      types.StringValue(item.Coffee.Teaser),
				Description: types.StringValue(item.Coffee.Description),
				Price:       types.Float64Value(item.Coffee.Price),
				Image:       types.StringValue(item.Coffee.Image),
			},
			Quantity: types.Int64Value(int64(item.Quantity)),
		})
	}
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *orderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state orderResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	err := o.client.DeleteOrder(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error Deleting HashiCups Order",
			"Could not delete order, unexpected error: "+err.Error(),
		)
		return
	}
}

func (o *orderResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	o.client = request.ProviderData.(*Client)
}

func (o *orderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
