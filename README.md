# Product API
 
A gRPC API for managing products.
 
---
 
## Base URL
 
```
/
```
 
---
 
## Endpoints
 
### Get Product by ID
 
```
GET /product/:id
```
 
Retrieves a single product by its unique identifier.
 
**Path Parameters**
 
| Parameter | Type     | Description              |
|-----------|----------|--------------------------|
| `id`      | `string` | The unique product ID    |
 
**Response**
 
```json
{
  "id": "123",
  "name": "Product Name",
  "price": 99.99
}
```
 
---
 
### Get Products by IDs
 
```
POST /fetch_by_ids
```
 
Retrieves multiple products by a list of IDs.
 
**Request Body**
 
```json
{
  "ids": ["123", "456", "789"]
}
```
 
**Response**
 
```json
[
  { "id": "123", "name": "Product A", "price": 10.00 },
  { "id": "456", "name": "Product B", "price": 20.00 }
]
```
 
---
 
### Add Product
 
```
POST /add
```
 
Creates and adds a new product.
 
**Request Body**
 
```json
{
  "name": "New Product",
  "price": 49.99,
  "description": "A great product"
}
```
 
**Response**
 
```json
{
  "id": "124",
  "name": "New Product",
  "price": 49.99,
  "description": "A great product"
}
```
 
---
 
### Calculate Total Price
 
```
POST /calculate_total_price
```
 
Calculates the total price for a given list of products and their quantities.
 
**Request Body**
 
```json
{
  "items": [
    { "id": "123", "quantity": 2 },
    { "id": "456", "quantity": 1 }
  ]
}
```
 
**Response**
 
```json
{
  "total": 40.00
}
```
 
---
 
### Update Product
 
```
PUT /product/:id
```
 
Updates an existing product by its ID.
 
**Path Parameters**
 
| Parameter | Type     | Description           |
|-----------|----------|-----------------------|
| `id`      | `string` | The unique product ID |
 
**Request Body**
 
```json
{
  "name": "Updated Name",
  "price": 59.99,
  "description": "Updated description"
}
```
 
**Response**
 
```json
{
  "id": "123",
  "name": "Updated Name",
  "price": 59.99,
  "description": "Updated description"
}
```
 
---
 
### Delete Product
 
```
DELETE /product/:id
```
 
Deletes a product by its ID.
 
**Path Parameters**
 
| Parameter | Type     | Description           |
|-----------|----------|-----------------------|
| `id`      | `string` | The unique product ID |
 
**Response**
 
```json
{
  "message": "Product deleted successfully"
}
```
 
---
 
## Route Summary
 
| Method   | Endpoint                  | Description                  |
|----------|---------------------------|------------------------------|
| `GET`    | `/product/:id`            | Get a product by ID          |
| `POST`   | `/fetch_by_ids`           | Get multiple products by IDs |
| `POST`   | `/add`                    | Add a new product            |
| `POST`   | `/calculate_total_price`  | Calculate total price        |
| `PUT`    | `/product/:id`            | Update a product by ID       |
| `DELETE` | `/product/:id`            | Delete a product by ID       |
 
---
 
## Error Responses
 
All endpoints may return the following error structure:
 
```json
{
  "error": "Description of the error"
}
```
 
| Status Code | Meaning               |
|-------------|-----------------------|
| `400`       | Bad Request           |
| `404`       | Product Not Found     |
| `500`       | Internal Server Error |
 
---
 
## Setup
 
```go
func ProductRoutes(r fiber.Router, c *controllers.Controller) {
    r.Get("/product/:id", c.GetProductByID)
    r.Post("/fetch_by_ids", c.GetProductsByIDs)
    r.Post("/add", c.AddProduct)
    r.Post("/calculate_total_price", c.CalculateTotalPrice)
    r.Put("/product/:id", c.UpdateProduct)
    r.Delete("/product/:id", c.DeleteProduct)
}
```
