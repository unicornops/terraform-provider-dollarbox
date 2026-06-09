package provider

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type memberModel struct {
	ID       types.String `tfsdk:"id"`
	Email    types.String `tfsdk:"email"`
	Role     types.String `tfsdk:"role"`
	JoinedAt types.String `tfsdk:"joined_at"`
}

func memberModelFromAPI(member apiMember) memberModel {
	return memberModel{
		ID:       stringValueOrNull(member.ID),
		Email:    stringValueOrNull(member.Email),
		Role:     stringValueOrNull(member.Role),
		JoinedAt: stringValueOrNull(member.JoinedAt),
	}
}

func updateMemberModelFromAPI(model *memberModel, member apiMember) {
	*model = memberModelFromAPI(member)
}

func stringValueOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func findMemberByEmail(members []apiMember, email string) (apiMember, bool) {
	email = strings.TrimSpace(email)
	for _, member := range members {
		if strings.EqualFold(member.Email, email) {
			return member, true
		}
	}
	return apiMember{}, false
}
