LawnConnect Backend API
 Project Overview
This is the backend for LawnConnect, a two-sided marketplace designed to connect homeowners with lawn care service providers. The API facilitates user management, booking services, and role-based access control, laying the foundation for a platform similar to on-demand services like Uber, but for lawn maintenance.

The platform's core objective is to handle the business logic that allows:

Customers to schedule and manage lawn care appointments.

Service Providers (Mowers) to accept job requests, set pricing, and manage their schedules.

 Key Features
User Authentication & Authorization: Secure user registration, login, and token-based authentication (JWT).

Role-Based Access Control: Distinguishes between customer and mower roles to restrict access to specific endpoints.

Booking Management: Customers can create, view, and cancel bookings. Mowers can accept, view, and complete bookings.

Dynamic Booking Status: Bookings have statuses such as pending, accepted, completed, and cancelled.

Simulated Payment: Mowers can set a price upon completing a job, and the API simulates a payment process.

Modular Architecture: Clean separation of concerns using a layered architecture (handlers, services, repositories).

 Technologies Used
Go: The core programming language for the backend.

Chi: A lightweight, idiomatic router for building REST APIs in Go.

MongoDB: A NoSQL database for storing user, booking, and other application data.

Cloudinary: Used for handling file uploads (e.g., user profile pictures).

SMTP: Used for sending transactional emails (e.g., password reset).

Godotenv: Manages environment variables for configuration.

Getting Started
Prerequisites
Go (version 1.18 or higher)

A running MongoDB instance (either local or cloud-based).

A Cloudinary account.

An SMTP server for sending emails.

Installation and Setup
Clone the repository:

git clone [\[your-repo-url\]](https://github.com/OsGift/lawnconnect-api)
cd lawnconnect-api

Set up environment variables:
Create a .env file in the root directory and populate it with your configuration details. A sample.env is provided to guide you.

MONGO_URI="mongodb://localhost:27017"
MONGO_DB_NAME="lawnconnect_db"
JWT_SECRET="your_secret_key"
CLOUDINARY_URL="cloudinary://your_api_key:your_api_secret@your_cloud_name"
SMTP_HOST="smtp.mailtrap.io"
SMTP_PORT=587
SMTP_USER="user"
SMTP_PASS="password"
FROM_EMAIL="noreply@lawnconnect.com"
TEMPLATES_PATH="./templates"
LOGIN_URL="http://localhost:8080/login"

Run the application:

go run main.go

The server will start on http://localhost:8080.

API Endpoints
All endpoints are prefixed with /api/v1.

Method

Endpoint

Description

Access Role

POST

/auth/register

Creates a new user account.

Public

POST

/auth/login

Authenticates a user and returns a JWT.

Public

POST

/bookings

Creates a new booking.

Customer

GET

/bookings

Retrieves all bookings for the authenticated user.

Both

GET

/bookings/{bookingID}

Retrieves a single booking by ID.

Both

PUT

/bookings/{bookingID}/cancel

Cancels an existing booking.

Customer

PUT

/bookings/{bookingID}/accept

Accepts a booking request.

Mower

PUT

/bookings/{bookingID}/complete

Completes a booking and sets the final price.

Mower

 Contributing
Feel free to submit issues or pull requests to improve the API's functionality, security, or performance.