# LawnConnect Backend API

## Overview

LawnConnect is a two-sided marketplace that connects homeowners with lawn care service providers. This backend API powers the core business logic—managing users, handling bookings, and enforcing role-based access—similar in concept to on-demand platforms like Uber, but tailored for lawn maintenance.

### Core Objectives

* **For Customers:** Schedule, track, and manage lawn care appointments.
* **For Service Providers (Mowers):** Accept job requests, set pricing, and manage schedules efficiently.

---

## Key Features

* **User Authentication & Authorization** – Secure registration, login, and JWT-based authentication.
* **Role-Based Access Control** – Separate permissions for customers and mowers to protect endpoints.
* **Booking Management** –

  * Customers: Create, view, and cancel bookings.
  * Mowers: Accept, view, and complete bookings.
* **Dynamic Booking Status** – Supports `pending`, `accepted`, `completed`, and `cancelled`.
* **Simulated Payments** – Mowers set the price after completing a job; payment is simulated for demo purposes.
* **Modular Architecture** – Clear separation of concerns using handlers, services, and repositories.

---

## Tech Stack

* **Go** – Primary backend language.
* **Chi** – Lightweight, idiomatic router for Go REST APIs.
* **MongoDB** – NoSQL database for user, booking, and application data.
* **Cloudinary** – Media storage for file uploads (e.g., profile images).
* **SMTP** – Transactional emails (e.g., password resets).
* **Godotenv** – Environment variable management.

---

## Getting Started

### Prerequisites

* Go **v1.18+**
* MongoDB (local or cloud)
* Cloudinary account
* SMTP server credentials

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/OsGift/lawnconnect-api
cd lawnconnect-api
```

2. **Set environment variables**
   Create a `.env` file in the root directory. Use `sample.env` as a guide:

```env
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
```

3. **Run the application**

```bash
go run main.go
```

The server will be available at `http://localhost:8080`.

---

## API Endpoints

All routes are prefixed with `/api/v1`.

| Method | Endpoint                         | Description                       | Role     |
| ------ | -------------------------------- | --------------------------------- | -------- |
| POST   | `/auth/register`                 | Register a new account            | Public   |
| POST   | `/auth/login`                    | Login and receive a JWT           | Public   |
| POST   | `/bookings`                      | Create a booking                  | Customer |
| GET    | `/bookings`                      | Get all bookings for current user | Both     |
| GET    | `/bookings/{bookingID}`          | Get booking by ID                 | Both     |
| PUT    | `/bookings/{bookingID}/cancel`   | Cancel a booking                  | Customer |
| PUT    | `/bookings/{bookingID}/accept`   | Accept a booking                  | Mower    |
| PUT    | `/bookings/{bookingID}/complete` | Complete a booking and set price  | Mower    |

---

## Contributing

Contributions are welcome! Please open issues or submit pull requests for improvements, bug fixes, or feature suggestions.
