package handlers

import (
	"backend/entities/author_type"
	"backend/entities/tender_status"
	"backend/storage"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"reflect"
	"regexp"
)

type Handlers struct {
	s         *storage.Storage
	validator *validator.Validate
}

func uidValidator(fl validator.FieldLevel) bool {
	switch v := fl.Field(); v.Kind() {
	case reflect.String:
		match, _ := regexp.Match("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
			[]byte(v.String()))
		return match
	default:
		return false
	}
}

func NewHandlers(s *storage.Storage) *Handlers {
	val := validator.New()
	val.RegisterValidation("uid", uidValidator)
	return &Handlers{
		s:         s,
		validator: val,
	}
}

func (h Handlers) Ping(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).SendString("ok")
}

type createTenderRequest struct {
	Name           string   `json:"name" validate:"required,max=50,min=5"`
	Description    string   `json:"description" validate:"required,max=1000,min=1"`
	ServiceType    []string `json:"serviceType" validate:"max=3,dive,oneof=Construction Delivery Manufacture"`
	OrganizationId string   `json:"organizationId" validate:"required,uid"`
}

func (h Handlers) CreateTender(c *fiber.Ctx) error {
	var request createTenderRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of body: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of body params: " + err.Error()})
	}
	tender, err := h.s.CreateTender(request.Name, request.Description, request.ServiceType, request.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(tender)
}

type filterTendersRequest struct {
	Limit       int      `json:"limit" validate:"min=0"`
	Offset      int      `json:"offset" validate:"min=0"`
	ServiceType []string `json:"serviceType" validate:"max=3,dive,oneof=Construction Delivery Manufacture"`
}

func (h Handlers) FilterTenders(c *fiber.Ctx) error {
	var request filterTendersRequest
	if err := c.QueryParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of body params: " + err.Error()})
	}
	request.Limit = c.QueryInt("limit", 5)
	request.Offset = c.QueryInt("offset", 0)
	tenders, err := h.s.FilterTenders(request.Limit, request.Offset, request.ServiceType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(tenders)
}

type filterMyTendersRequest struct {
	Limit    int    `json:"limit" validate:"min=0"`
	Offset   int    `json:"offset" validate:"min=0"`
	Username string `json:"username" validate:"required,max=50"`
}

func (h Handlers) FilterMyTenders(c *fiber.Ctx) error {
	var request filterMyTendersRequest
	if err := c.QueryParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query params: " + err.Error()})
	}
	request.Limit = c.QueryInt("limit", 5)
	request.Offset = c.QueryInt("offset", 0)
	userId, err := h.s.GetUserId(request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	tenders, err := h.s.FilterUsersTenders(request.Limit, request.Offset, userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(tenders)
}

type getTenderStatusRequest struct {
	Username string `json:"username" validate:"max=50"`
}

func (h Handlers) GetTenderStatus(c *fiber.Ctx) error {
	var request getTenderStatusRequest
	if err := c.QueryParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query params: " + err.Error()})
	}
	tenderId := c.Params("tenderId")
	tender, err := h.s.GetTender(tenderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Tender not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if tender.Status == tender_status.PUBLISHED {
		return c.Status(fiber.StatusOK).SendString(tender_status.PUBLISHED)
	}
	userId, err := h.s.GetUserId(request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	checkPermission, err := h.s.CheckOrganizationResponsible(userId, tender.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if !checkPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"reason": "user has no permission to see this tender"})
	}
	return c.Status(fiber.StatusOK).SendString(tender.Status)
}

type updateStatusRequest struct {
	Status   string `json:"status" validate:"required,oneof=Created Published Closed"`
	Username string `json:"username" validate:"required,max=50"`
}

func (h Handlers) UpdateTenderStatus(c *fiber.Ctx) error {
	var request updateStatusRequest
	if err := c.QueryParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query params: " + err.Error()})
	}
	userId, err := h.s.GetUserId(request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	tenderId := c.Params("tenderId")
	tender, err := h.s.GetTender(tenderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Tender is not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	checkPermission, err := h.s.CheckOrganizationResponsible(userId, tender.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if !checkPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"reason": "user has no permission to see this tender"})
	}
	tenderNew, err := h.s.PatchTender(tenderId, nil, nil, &request.Status, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(tenderNew)
}

type editTenderRequest struct {
	Name        string   `json:"name,omitempty" validate:"omitempty,max=50"`
	Description string   `json:"description,omitempty" validate:"omitempty,max=1000,min=1"`
	ServiceType []string `json:"serviceType,omitempty" validate:"omitempty,max=3,dive,oneof=Construction Delivery Manufacture"`
	Status      string   `json:"status,omitempty" validate:"omitempty,oneof=Created Published Closed"`
}

func (h Handlers) EditTender(c *fiber.Ctx) error {
	var request editTenderRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query params: " + err.Error()})
	}
	username := c.Query("username")
	fmt.Println("+" + username + "+")
	userId, err := h.s.GetUserId(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	tenderId := c.Params("tenderId")
	tender, err := h.s.GetTender(tenderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Tender is not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	checkPermission, err := h.s.CheckOrganizationResponsible(userId, tender.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if !checkPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"reason": "user has no permission to see this tender"})
	}

	var name *string = nil
	var description *string = nil
	var serviceType []string = nil
	var status *string = nil
	if len(request.Name) > 0 {
		name = &request.Name
	}
	if len(request.Description) > 0 {
		description = &request.Description
	}
	if request.ServiceType != nil {
		serviceType = request.ServiceType
	}
	if len(request.Status) > 0 {
		status = &request.Status
	}
	tenderNew, err := h.s.PatchTender(tenderId, name, description, status, serviceType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(tenderNew)
}

type createBidRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	Description string `json:"description" validate:"required,max=1000,min=1"`
	TenderId    string `json:"tenderId" validate:"required,uid"`
	AuthorType  string `json:"authorType" validate:"required,oneof=User Organization"`
	AuthorId    string `json:"authorId" validate:"required,uid"`
}

func (h Handlers) CreateBid(c *fiber.Ctx) error {
	var request createBidRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of body: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of body params: " + err.Error()})
	}
	if request.AuthorType == author_type.ORGANIZATION {
		_, err := h.s.GetOrganization(request.AuthorId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Organization is not found: " + err.Error()})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
		}
	} else {
		_, err := h.s.GetUser(request.AuthorId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
		}
	}
	_, err := h.s.GetTender(request.TenderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Tender is not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	bid, err := h.s.CreateBid(request.Name, request.Description, request.AuthorType, request.AuthorId, request.TenderId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(bid)
}

type filterBidsRequest struct {
	Limit    int    `json:"limit" validate:"min=0"`
	Offset   int    `json:"offset" validate:"min=0"`
	Username string `json:"username" validate:"max=50"`
}

func (h Handlers) GetMyBids(c *fiber.Ctx) error {
	var request filterBidsRequest
	if err := c.QueryParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query params: " + err.Error()})
	}
	request.Limit = c.QueryInt("limit", 5)
	request.Offset = c.QueryInt("offset", 0)
	userId, err := h.s.GetUserId(request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	bids, err := h.s.GetMyBids(userId, request.Limit, request.Offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(bids)
}

type filterBidsByTenderRequest struct {
	Limit    int    `json:"limit" validate:"min=0"`
	Offset   int    `json:"offset" validate:"min=0"`
	Username string `json:"username" validate:"required,max=50"`
}

func (h Handlers) GetTenderBids(c *fiber.Ctx) error {
	var request filterBidsByTenderRequest
	if err := c.QueryParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query params: " + err.Error()})
	}
	request.Limit = c.QueryInt("limit", 5)
	request.Offset = c.QueryInt("offset", 0)
	tenderId := c.Params("tenderId")
	userId, err := h.s.GetUserId(request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	tender, err := h.s.GetTender(tenderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Tender is not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	permission, err := h.s.CheckOrganizationResponsible(userId, tender.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if !permission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"reason": "User has no permission to see these bids"})
	}
	fmt.Println("+"+tenderId+"+", request.Limit, request.Offset)
	bids, err := h.s.GetBidsByTender(tenderId, request.Limit, request.Offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(bids)
}

type getBidStatusRequest struct {
	Username string `json:"username" validate:"max=50,required"`
}

func (h Handlers) GetBidStatus(c *fiber.Ctx) error {
	var request getBidStatusRequest
	if err := c.QueryParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query params: " + err.Error()})
	}
	bidId := c.Params("bidId")
	userId, err := h.s.GetUserId(request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	bid, err := h.s.GetBid(bidId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Bid is not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}

	perm := false
	tender, err := h.s.GetTender(bid.TenderId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	ok, err := h.s.CheckOrganizationResponsible(userId, tender.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if ok {
		perm = true
	}
	if bid.AuthorType == author_type.USER && bid.AuthorId == userId {
		perm = true
	}
	if !perm {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"reason": "User has no permission to see this bid"})
	}

	return c.Status(fiber.StatusOK).SendString(bid.Status)
}

type changeBidStatusRequest struct {
	Status   string `json:"status" validate:"required,oneof=Created Published Cancelled"`
	Username string `json:"username" validate:"max=50,required"`
}

func (h Handlers) ChangeBidStatus(c *fiber.Ctx) error {
	var request changeBidStatusRequest
	if err := c.QueryParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query params: " + err.Error()})
	}
	bidId := c.Params("bidId")
	userId, err := h.s.GetUserId(request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	bid, err := h.s.GetBid(bidId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Bid is not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}

	perm := false
	tender, err := h.s.GetTender(bid.TenderId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	ok, err := h.s.CheckOrganizationResponsible(userId, tender.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if ok {
		perm = true
	}
	if bid.AuthorType == author_type.USER && bid.AuthorId == userId {
		perm = true
	}
	if !perm {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"reason": "User has no permission to see this bid"})
	}

	bidNew, err := h.s.PatchBid(bidId, nil, nil, &request.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(bidNew)
}

type editBidRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,max=50"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000,min=1"`
	Status      string `json:"status,omitempty" validate:"omitempty,oneof=Created Published Closed"`
}

func (h Handlers) EditBid(c *fiber.Ctx) error {
	var request editBidRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of body: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of body params: " + err.Error()})
	}
	bidId := c.Params("bidId")
	bid, err := h.s.GetBid(bidId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Bid is not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	username := c.Query("username")
	userId, err := h.s.GetUserId(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}

	perm := false
	tender, err := h.s.GetTender(bid.TenderId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	ok, err := h.s.CheckOrganizationResponsible(userId, tender.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if ok {
		perm = true
	}
	if bid.AuthorType == author_type.USER && bid.AuthorId == userId {
		perm = true
	}
	if !perm {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"reason": "User has no permission to see this bid"})
	}

	var name *string = nil
	var description *string = nil
	var status *string = nil
	if len(request.Name) > 0 {
		name = &request.Name
	}
	if len(request.Description) > 0 {
		description = &request.Description
	}
	if len(request.Status) > 0 {
		status = &request.Status
	}
	newBid, err := h.s.PatchBid(bidId, name, description, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(newBid)
}

type setDecisionRequest struct {
	Decision string `json:"decision" validate:"oneof=Approved Rejected"`
	Username string `json:"username" validate:"max=50"`
}

func (h Handlers) SetDecision(c *fiber.Ctx) error {
	var request setDecisionRequest
	if err := c.QueryParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query: " + err.Error()})
	}
	if err := h.validator.Struct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"reason": "Wrong format of query params: " + err.Error()})
	}
	bidId := c.Params("bidId")
	bid, err := h.s.GetBid(bidId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Bid is not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	userId, err := h.s.GetUserId(request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}

	tender, err := h.s.GetTender(bid.TenderId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	permission, err := h.s.CheckOrganizationResponsible(userId, tender.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if !permission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"reason": "User has no permission to see this bid"})
	}

	err = h.s.SetDecision(bidId, request.Decision)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	newStatus := tender_status.CLOSED
	tenderNew, _ := h.s.PatchTender(tender.Id, nil, nil, &newStatus, nil)
	return c.Status(fiber.StatusOK).JSON(tenderNew)
}

func (h Handlers) GetDecision(c *fiber.Ctx) error {
	bidId := c.Params("bidId")
	bid, err := h.s.GetBid(bidId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"reason": "Bid is not found: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	username := c.Query("username")
	userId, err := h.s.GetUserId(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"reason": "User is not correct: " + err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}

	perm := false
	tender, err := h.s.GetTender(bid.TenderId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	ok, err := h.s.CheckOrganizationResponsible(userId, tender.OrganizationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	if ok {
		perm = true
	}
	if bid.AuthorType == author_type.USER && bid.AuthorId == userId {
		perm = true
	}
	if !perm {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"reason": "User has no permission to see this bid"})
	}

	decision, err := h.s.GetDecision(bidId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"reason": "internal server error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).SendString(decision)
}
