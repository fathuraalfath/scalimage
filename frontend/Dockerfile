FROM node:24-alpine

WORKDIR /app

# Cache package dependencies
COPY package*.json ./
RUN npm install

# Copy source files
COPY . .

# Expose Vite port
EXPOSE 5173

# Run Vite dev server, exposing it to external container networks
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
