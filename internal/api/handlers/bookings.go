package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"lawnconnect-api/internal/core/services"
	httpresponse "lawnconnect-api/internal/api/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/go-chi/chi/v5"
	"lawnconnect-api/internal/core/apperror"
)

// BookingHandler handles HTTP requests related to bookings.
type BookingHandler struct {
	BookingService services.BookingService
}

// NewBookingHandler creates a new BookingHandler.
func NewBookingHandler(bookingSrv services.BookingService) *BookingHandler {
	return &BookingHandler{BookingService: bookingSrv}
}

// CreateBooking handles creating a new booking.
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		CustomerID  string `json:"customerId"`
		Date        string `json:"date"`
		Time        string `json:"time"`
		Address     string `json:"address"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	customerID, err := primitive.ObjectIDFromHex(reqBody.CustomerID)
	if err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	booking, err := h.BookingService.CreateBooking(r.Context(), customerID, reqBody.Date, reqBody.Time, reqBody.Address, reqBody.Description)
	if err != nil {
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to create booking")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusCreated, "Booking created successfully", booking)
}

// GetBookingByID retrieves a single booking by its ID.
func (h *BookingHandler) GetBookingByID(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingID")
	bookingID, err := primitive.ObjectIDFromHex(bookingIDStr)
	if err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	booking, err := h.BookingService.GetBookingByID(r.Context(), bookingID)
	if err != nil {
		log.Printf("Error getting booking: %v", err)
		httpresponse.JSONError(w, http.StatusNotFound, "Booking not found")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Booking retrieved successfully", booking)
}

// AcceptBooking handles a mower accepting a booking.
func (h *BookingHandler) AcceptBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingID")
	bookingID, err := primitive.ObjectIDFromHex(bookingIDStr)
	if err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	// Get mower ID from context
	mowerID := r.Context().Value(UserContextKey).(primitive.ObjectID)

	err = h.BookingService.AcceptBooking(r.Context(), bookingID, mowerID)
	if err != nil {
		log.Printf("Error accepting booking: %v", err)
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to accept booking")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Booking accepted successfully", nil)
}

// CompleteBooking handles a mower completing a booking, setting the final price.
func (h *BookingHandler) CompleteBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingID")
	bookingID, err := primitive.ObjectIDFromHex(bookingIDStr)
	if err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	var reqBody struct {
		Price float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if reqBody.Price <= 0 {
		httpresponse.JSONError(w, http.StatusBadRequest, "Price must be a positive number")
		return
	}

	err = h.BookingService.CompleteBooking(r.Context(), bookingID, reqBody.Price)
	if err != nil {
		if _, ok := err.(apperror.CustomError); ok {
			httpresponse.JSONError(w, http.StatusConflict, err.Error())
			return
		}
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to complete booking")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Booking completed and payment simulated successfully", nil)
}

// ListBookings retrieves a list of bookings for the authenticated user.
func (h *BookingHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserContextKey).(primitive.ObjectID)

	bookings, err := h.BookingService.ListBookings(r.Context(), userID)
	if err != nil {
		log.Printf("Error listing bookings: %v", err)
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to retrieve bookings")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Bookings retrieved successfully", bookings)
}

// CancelBooking handles a customer cancelling their booking.
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingID")
	bookingID, err := primitive.ObjectIDFromHex(bookingIDStr)
	if err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	customerID := r.Context().Value(UserContextKey).(primitive.ObjectID)

	err = h.BookingService.CancelBooking(r.Context(), bookingID, customerID)
	if err != nil {
		if _, ok := err.(apperror.CustomError); ok {
			httpresponse.JSONError(w, http.StatusForbidden, err.Error())
			return
		}
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to cancel booking")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Booking cancelled successfully", nil)
}

// CustomerRoutes mounts the booking-related routes for customers to a chi router.
func (h *BookingHandler) CustomerRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListBookings) // GET /api/v1/bookings
	r.Post("/", h.CreateBooking) // POST /api/v1/bookings
	r.Get("/{bookingID}", h.GetBookingByID) // GET /api/v1/bookings/{bookingID}
	r.Put("/{bookingID}/cancel", h.CancelBooking) // PUT /api/v1/bookings/{bookingID}/cancel
	return r
}

// MowerRoutes mounts the booking-related routes for mowers to a chi router.
func (h *BookingHandler) MowerRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListBookings) // GET /api/v1/bookings
	r.Put("/{bookingID}/accept", h.AcceptBooking) // PUT /api/v1/bookings/{bookingID}/accept
	r.Put("/{bookingID}/complete", h.CompleteBooking) // PUT /api/v1/bookings/{bookingID}/complete
	return r
}
