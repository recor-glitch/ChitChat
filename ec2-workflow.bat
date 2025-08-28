#! /bin/bash

sudo yum update -y
sudo yum install -y docker

sudo systemctl start docker
sudo systemctl enable docker

sudo docker pull rishi513/chitchat:latest

sudo docker run -d -p 5005:5005 rishi513/chitchat:latest

sudo yum install -y nginx

sudo systemctl start nginx
sudo systemctl enable nginx

cat > nginx.conf <<'EOF'
events {}

http {
  server {
    listen 80;
    server_name _;

    location / {
      proxy_pass http://chitchat_backend:4000;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }
  }
}
EOF

sudo mv nginx.conf /etc/nginx/nginx.conf
sudo systemctl restart nginx