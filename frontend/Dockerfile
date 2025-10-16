FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install ALL dependencies (devDependencies needed for Vite build)
RUN npm ci

# Copy source code
COPY . .

# Accept build arguments
ARG VITE_API_BASE_URL
ARG VITE_GOOGLE_CLIENT_ID
ARG VITE_VERSION=1.0.0
ARG VITE_GIT_COMMIT=unknown
ARG VITE_BUILD_ENV=development

# Set environment variables for build
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID
ENV VITE_VERSION=$VITE_VERSION
ENV VITE_GIT_COMMIT=$VITE_GIT_COMMIT
ENV VITE_BUILD_ENV=$VITE_BUILD_ENV

# Build the application
RUN npm run build

FROM nginx:alpine

# Copy built application
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy startup script
COPY start.sh /start.sh
RUN chmod +x /start.sh

# Expose port
EXPOSE 8080

# Start nginx with dynamic port
CMD ["/start.sh"]
