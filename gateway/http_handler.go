package main

import (
	"errors"
	"net/http"

	common "github.com/salvatoreolivieri/commons"
	pb "github.com/salvatoreolivieri/commons/api"
	"github.com/salvatoreolivieri/omsv-gateway/gateway"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	gateway gateway.OrdersGateway
}

type ItemsQuantity = []*pb.ItemsWithQuantity

func NewHandler(gateway gateway.OrdersGateway) *handler {
	return &handler{gateway}
}

func (h *handler) registerRoutes(mux *http.ServeMux) {
	// static folder serving
	mux.Handle("/", http.FileServer(http.Dir("public")))

	mux.HandleFunc("POST /api/customers/{customerID}/orders", h.handleCreateOrder)
	mux.HandleFunc("GET /api/customers/{customerID}/orders/{orderID}", h.handleGetOrder)
}

func (h *handler) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")
	orderID := r.PathValue("orderID")

	order, err := h.gateway.GetOrder(r.Context(), orderID, customerID)

	responseStatus := status.Convert(err)

	if responseStatus != nil {

		if responseStatus.Code() != codes.InvalidArgument {
			common.WriteError(w, http.StatusBadRequest, responseStatus.Message())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.WriteJSON(w, http.StatusOK, order)
}

func (h *handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")

	var items ItemsQuantity
	if err := common.ReadJSON(w, r, &items); err != nil {
		common.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := validateItems(items)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.gateway.CreateOrder(r.Context(), &pb.CreateOrderRequest{
		CustomerID: customerID,
		Items:      items,
	})

	responseStatus := status.Convert(err)

	if responseStatus != nil {

		if responseStatus.Code() != codes.InvalidArgument {
			common.WriteError(w, http.StatusBadRequest, responseStatus.Message())
			return
		}

		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.WriteJSON(w, http.StatusOK, order)
}

func validateItems(items ItemsQuantity) error {
	if len(items) == 0 {
		return common.ErrNoItems
	}

	for _, item := range items {
		if item.ID == "" {
			return errors.New("item ID is required")
		}

		if item.Quantity <= 0 {
			return errors.New("item must have a valid quantity")
		}
	}

	return nil
}
