# uar

## Running it locally
### Prerequisites
- go 1.23.2

### Installation
1. Clone the repo
   ```sh
   git clone https://github.com/praktykanci-IBM/uar-cli-tool
   ```
2. Set your GitHub PAT in `.env`
   ```sh
   cat .enx.example > .env
   ```
   ```
   # .env
   GITHUB_PAT="your_token"
   ```
   
### Usage
To use uar you can either:
- run it without saving binary (recommened for developing)
  ```sh
  go run .
  ```
- or compile it to the binary, and run it
  ```
  go build
  ./uar
  ```
