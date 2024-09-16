package entities

type OrganizationResponsible struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organizationId"`
	UserId         string `json:"userId"`
}
