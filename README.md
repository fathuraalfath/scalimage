# Scalimage 🛠️🖼️

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-19.x-61DAFB?style=flat&logo=react)](https://reactjs.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> A scalable, high-performance web application for generating image collages and managing versatile image processing utilities.

## 📖 About The Project

Scalimage is designed as a powerful internal utility for manipulating and combining images, built with an eye toward open-source self-hosting and distribution. Starting as a robust image collage generator, its modular Golang backend and interactive React.js frontend are architected to scale. The roadmap includes supporting a full suite of image utilities such as compression, resizing, and format conversion.

## ✨ Features

### 🚀 Current Capabilities
* **Collage Maker**: Upload multiple images, arrange them using dynamic pre-defined paper-size grids (A4, F4, A5, Square) with portrait/landscape toggles, and export the high-resolution output.
* **Auto-Filling Layouts**: Auto-generates slot layouts mapping 1-to-1 with the number of uploaded images (from 1 to 8).
* **High-Performance Processing**: Leverages Go's concurrency model (goroutines) to decode and scale images in parallel.
* **Directory Traversal Guard**: Secure upload storage handling with validation checks on absolute and root-relative paths.
* **Docker Support**: Ready to launch in a unified environment with Docker Compose.

---

## 🛠️ Tech Stack

* **Backend**: [Golang](https://go.dev/) (Standard Library, official sub-repo Image scaling utilities)
* **Frontend**: [React.js](https://reactjs.org/) (Vite, Vanilla CSS)
* **Containerization**: [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)

---

## 🚀 Getting Started

To get a local copy up and running, choose one of the setup options below.

### Option A: Running with Docker (Recommended)

If you have Docker and Docker Compose installed, you can build and run both the frontend and backend with a single command:

```sh
docker compose up --build
```
* **Frontend Web App**: `http://localhost:5173`
* **Backend API Server**: `http://localhost:8080`
* Uploaded files will persist in a named Docker volume (`scalimage-upload-data`).

---

### Option B: Local Setup

#### Prerequisites
* [Go](https://go.dev/doc/install) (v1.21 or higher)
* [Node.js](https://nodejs.org/) (v18 or higher)
* npm

#### 1. Clone the repository
```sh
git clone https://github.com/fathuraalfath/scalimage.git
cd scalimage
```

#### 2. Run the Backend (Golang)
```sh
cd backend
go mod tidy
go run cmd/server/main.go
```
*The backend server will start on `http://localhost:8080`*

#### 3. Run the Frontend (React + Vite)
```sh
cd ../frontend
npm install
npm run dev
```
*The frontend development server will start on `http://localhost:5173`*

---

### 🧪 Running Tests

To run unit and integration tests for the Go backend:

```sh
cd backend
go test -v ./...
```

---

## 🤝 Contributing

Since this project is intended for open-source, contributions are highly encouraged! If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.