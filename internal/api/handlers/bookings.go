package handlers

import (
	"encoding/json"
	httpresponse "lawnconnect-api/internal/api/http"
	"lawnconnect-api/internal/core/apperror"
	"lawnconnect-api/internal/core/services"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		Date        string `json:"date"`
		Time        string `json:"time"`
		Address     string `json:"address"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	customerID := r.Context().Value(UserContextKey).(primitive.ObjectID)

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
		if _, ok := err.(apperror.NotFound); ok {
			httpresponse.JSONError(w, http.StatusNotFound, "Booking not found")
			return
		}
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to retrieve booking")
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

// ListPendingBookings handles listing all pending bookings for mowers.
func (h *BookingHandler) ListPendingBookings(w http.ResponseWriter, r *http.Request) {
	bookings, err := h.BookingService.ListPendingBookings(r.Context())
	if err != nil {
		log.Printf("Error listing pending bookings: %v", err)
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to retrieve pending bookings")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Pending bookings retrieved successfully", bookings)
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

// RejectBooking handles a mower rejecting a booking.
func (h *BookingHandler) RejectBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingID")
	bookingID, err := primitive.ObjectIDFromHex(bookingIDStr)
	if err != nil {
		httpresponse.JSONError(w, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	mowerID := r.Context().Value(UserContextKey).(primitive.ObjectID)

	err = h.BookingService.RejectBooking(r.Context(), bookingID, mowerID)
	if err != nil {
		if _, ok := err.(apperror.CustomError); ok {
			httpresponse.JSONError(w, http.StatusForbidden, err.Error())
			return
		}
		httpresponse.JSONError(w, http.StatusInternalServerError, "Failed to reject booking")
		return
	}

	httpresponse.JSONSuccess(w, http.StatusOK, "Booking rejected successfully", nil)
}

// Routes mounts the booking-related routes to a chi router.
func (h *BookingHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.CreateBooking)                      // POST /api/v1/bookings
	r.Get("/", h.ListBookings)                        // GET /api/v1/bookings
	r.Get("/pending", h.ListPendingBookings)          // GET /api/v1/bookings/pending
	r.Get("/{bookingID}", h.GetBookingByID)           // GET /api/v1/bookings/{bookingID}
	r.Put("/{bookingID}/accept", h.AcceptBooking)     // PUT /api/v1/bookings/{bookingID}/accept
	r.Put("/{bookingID}/complete", h.CompleteBooking) // PUT /api/v1/bookings/{bookingID}/complete
	r.Put("/{bookingID}/cancel", h.CancelBooking)     // PUT /api/v1/bookings/{bookingID}/cancel
	r.Put("/{bookingID}/reject", h.RejectBooking)     // PUT /api/v1/bookings/{bookingID}/reject

	return r
}
