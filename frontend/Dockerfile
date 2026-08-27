# Stage 1: Build production static bundle
FROM node:24-alpine AS builder

WORKDIR /app

# Cache package dependencies
COPY package*.json ./
RUN npm install

# Copy source files and compile minified production assets
COPY . .
RUN npm run build

# Stage 2: Serve with lightweight Nginx web server (~25MB)
FROM nginx:alpine

# Copy custom Nginx SPA routing config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy compiled static assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose HTTP port
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
