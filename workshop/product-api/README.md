# Develope REST API
* [Echo](https://echo.labstack.com/)

## 0. Requirements
```
GET product/:id

code=200
{
    "id": 1,
    "name": "Product name 1",
    "price": 100.50,
    "stock": 10
}

code=404
{
    "message": "product id=2 not found in system"
}
```

## Create project
```
$go mod init demo 
```

## Create product model
* models/product.go

## Generate swagger
```
$go get github.com/swaggo/swag
$swag init
```

URL of swagger
* http://localhost:8080/swagger/index.html


