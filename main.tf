provider "aws" {
  region = "us-east-1" # Choose your preferred AWS region
}

# Create a security group for RDS
resource "aws_security_group" "rds_sg" {
  name        = "rds-security-group"
  description = "Allow inbound PostgreSQL traffic"

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # Allow access from anywhere (for demonstration purposes)
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Create an RDS PostgreSQL instance
resource "aws_db_instance" "postgresql" {
  identifier           = "logmonitoringdb"
  engine               = "postgres"
  engine_version       = "13.7" # PostgreSQL version
  instance_class       = "db.t3.micro" # Instance type
  allocated_storage    = 20 # Storage in GB
  storage_type         = "gp2"
  username             = "admin" # Database username
  password             = "password123" # Database password
  publicly_accessible  = true # Make the database publicly accessible (for demonstration purposes)
  skip_final_snapshot  = true # Do not create a snapshot upon deletion
  vpc_security_group_ids = [aws_security_group.rds_sg.id]
}

# Output the database connection URL
output "rds_endpoint" {
  value = "rds-endpoint.us-east-1.rds.amazonaws.com"
}