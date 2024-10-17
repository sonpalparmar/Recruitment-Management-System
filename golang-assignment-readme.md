# GolangAssignment

GolangAssignment is a web-based application that is built using Go and the Gin framework for the backend.

## Table of Contents
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Running the Application](#running-the-application)
- [API Endpoints](#api-endpoints)
- [Database Schema](#database-schema)
- [Technologies Used](#technologies-used)
- [Project Architecture](#project-architecture)
- [Contributing](#contributing)
- [License](#license)

## Features
- **Resume Upload**: Users can upload resumes in PDF or DOCX format.
- **Resume Parsing**: The system parses the resumes to extract information such as name, education, experience, skills, and contact details.
- **Database Integration**: The parsed data is stored in a PostgreSQL database.
- **Secure**: Authentication and authorization using JWT tokens.

## Prerequisites
Ensure you have the following installed:
- Go
- PostgreSQL
- Docker (optional for containerization)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/GolangAssignment.git
   cd GolangAssignment
   ```

2. Set up environment variables:
   Create a `.env` file in the root of your project with the following details:
   ```bash
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_NAME=your_db_name
   DB_HOST=localhost
   DB_PORT=5432
   API_KEY=api_key
   ```

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Set up PostgreSQL database:
   Create a database and ensure your connection details match the ones in `.env`.

5. Run database migrations:
   ```bash
   go run cmd/app/main.go
   ```

## Running the Application

To start the application, use the following command:

```bash
go run cmd/app/main.go
```

The server will start at `localhost:8080`.

## API Endpoints

### POST /uploadResume
Uploads a resume and parses it.

**Request:**
```bash
POST /uploadResume
Headers:
  - Content-Type: multipart/form-data
  - Authorization: Bearer <token>
Body:
  - resume: <file.pdf or file.docx>
```

**Response:**
```json
{
  "status": "success",
  "message": "Resume uploaded and parsed successfully",
  "data": {
    "name": "John Doe",
    "email": "john.doe@example.com",
    "education": "Bachelor's in Computer Science",
    "skills": "Golang, Docker, PostgreSQL",
    "experience": "Software Engineer at XYZ"
  }
}
```

## Database Schema

### Users Table

| Column        | Type      | Description              |
|---------------|-----------|--------------------------|
| id            | bigserial | Primary Key              |
| name          | text      | User's name              |
| email         | text      | User's email             |
| address       | text      | User's address           |
| user_type     | varchar(10)| Role (e.g., admin, applicant) |
| password_hash | text      | Hashed password          |

### Profiles Table

| Column     | Type      | Description              |
|------------|-----------|--------------------------|
| id         | bigserial | Primary Key              |
| user_id    | bigint    | Foreign Key (users)      |
| education  | text      | User's education         |
| experience | text      | User's experience        |
| skills     | text      | User's skills            |
| resume_file| text      | Path to resume file      |

## Project Architecture

```
├── cmd
│   └── app
│       └── main.go           # Entry point of the application
├── internal
│   ├── controllers
│   │   └── applicant_controller.go  # Handles resume upload and parsing
│   ├── models
│   │   └── user.go           # User and profile models
│   └── middlewares
│       └── auth_middleware.go  # JWT Authentication
├── config
│   └── config.go             # Configuration handling
├── pkg
│   └── utils
│       └── jwt.go            # JWT token utilities
└── .env                      # Environment variables


## Technologies Used
- **Go**: Programming language
- **Gin**: Web framework
- **PostgreSQL**: Database
- **Docker**: For containerization
- **JWT**: Authentication tokens

## Contributing

Contributions are welcome! Please create a pull request or open an issue for suggestions or bug reports.

## License

[Add your chosen license here]
