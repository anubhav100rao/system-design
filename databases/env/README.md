# Environment Variables for MongoDB

This directory contains the `.env` file for storing sensitive MongoDB credentials and configuration. 

## Example `.env` file
```
MONGODB_USERNAME=your_username
MONGODB_PASSWORD=your_password
MONGODB_HOST=localhost
MONGODB_PORT=27017
MONGODB_DATABASE=your_database
```

## Usage in Go
To use these variables in your Go application, you need to load them at runtime. The recommended package is [`github.com/joho/godotenv`](https://github.com/joho/godotenv).

## Install godotenv
```sh
cd ../..
cd databases
go get github.com/joho/godotenv
```
