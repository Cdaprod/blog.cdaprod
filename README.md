# Blog.cdaprod

A scalable blog application built with Go, MinIO for storage, Docker for containerization, and Cloudflare for secure HTTPS and DNS management.

## Features

- **Go**: Backend server using the Go programming language.
- **MinIO**: Object storage for storing blog posts.
- **Docker**: Containerization of the application.
- **Cloudflare**: Secure HTTPS and DNS management.
- **Caddy**: Automatic HTTPS configuration.

## Prerequisites

- [Go](https://golang.org/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [MinIO](https://min.io/)
- [Caddy](https://caddyserver.com/docs/install)
- [Cloudflare](https://www.cloudflare.com/) account

## Project Structure

```
blog.cdaprod/
├── main.go
├── database.go
├── templates/
│   ├── index.html
│   ├── post.html
│   └── create.html
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## Installation

### 1. Clone the repository

```sh
git clone https://github.com/Cdaprod/blog.cdaprod.git
cd blog.cdaprod
```

### 2. Set Up MinIO

Ensure you have MinIO running and create a bucket named `blog-posts`. Update your MinIO credentials in `main.go`.

### 3. Build and Run the Docker Container

Build the Docker image:

```sh
docker build -t blog.cdaprod .
```

Run the Docker container:

```sh
docker run -p 8080:8080 blog.cdaprod
```

### 4. Set Up HTTPS with Cloudflare

- Create an `A` record for `blog.cdaprod` pointing to your server’s IP address in Cloudflare.
- Enable SSL/TLS in "Full (strict)" mode in Cloudflare.

### 5. Set Up Caddy for Automatic HTTPS

Create a `Caddyfile`:

```caddyfile
blog.cdaprod {
    reverse_proxy localhost:8080
}
```

Start Caddy:

```sh
caddy start --config Caddyfile
```

## Usage

### Access the Application

Visit `https://blog.cdaprod` in your web browser.

### Creating a Post

1. Navigate to `/create`.
2. Fill in the form with your post details.
3. Submit the form to create a new post.

### Viewing Posts

1. Navigate to the home page to see a list of posts.
2. Click on a post title to view the full post.

## Code Overview

### main.go

- Sets up the HTTP server and routes.
- Handles requests to create, view, and list blog posts.
- Interacts with MinIO to store and retrieve blog content.

### database.go

- Initializes the SQLite database.
- Sets up the `posts` table for storing post metadata.

### templates/

Contains HTML templates for the application:
- `index.html`: Home page.
- `post.html`: Post detail page.
- `create.html`: Create post form.

### Dockerfile

- Defines the Docker image for the application.
- Specifies environment variables for MinIO credentials.

## Environment Variables

Ensure the following environment variables are set for MinIO:

```sh
MINIO_ENDPOINT=play.min.io
MINIO_ACCESS_KEY=YOUR-ACCESS-KEY
MINIO_SECRET_KEY=YOUR-SECRET-KEY
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Make your changes.
4. Commit your changes (`git commit -m 'Add some feature'`).
5. Push to the branch (`git push origin feature-branch`).
6. Open a pull request.

## Acknowledgements

- [Go](https://golang.org/)
- [MinIO](https://min.io/)
- [Docker](https://www.docker.com/)
- [Cloudflare](https://www.cloudflare.com/)
- [Caddy](https://caddyserver.com/)

## Contact

David Cannan - [LinkedIn](https://www.linkedin.com/in/cdasmkt) - [Twitter](https://twitter.com/cdasmktcda)

Feel free to reach out for any questions or collaboration opportunities!
```

This `README.md` covers the project overview, installation steps, usage instructions, code structure, environment variables, and contribution guidelines. Make sure to replace placeholders like `YOUR-ACCESS-KEY` and `YOUR-SECRET-KEY` with your actual credentials before running the application.