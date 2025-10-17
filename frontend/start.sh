#!/bin/sh

# Get the port from environment variable or default to 8080
PORT=${PORT:-8080}

# Replace the port in nginx configuration
sed -i "s/listen       80;/listen       ${PORT};/g" /etc/nginx/nginx.conf

# Start nginx
nginx -g "daemon off;"
