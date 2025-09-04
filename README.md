# Pismo Phase 1 API

## Run
docker compose up --build

API: http://localhost:8080
Swagger UI: http://localhost:8080/docs

## Endpoints
POST /accounts
{
  "document_number": "12345678900"
}

GET /accounts?account_id=1
GET /accounts/1

POST /transactions
{
  "account_id": 1,
  "operation_type_id": 1,   // 1=purchase, 2=installment, 3=withdrawal, 4=payment
  "amount": 123.45          // debits are saved as negative for 1-3, positive for 4
}
