# Document DB in Go

I created the document DB referring to [the blog](https://notes.eatonphil.com/documentdb.html).

## Usage

```sh
$ go run main.go

$ curl -X POST \
    -H 'Content-Type: application/json' \
    -d '{"id": "1", "name": "bookA", "detail": {"price": 100,"description": "this is sample book"}}' \
    http://localhost:8080/docs
{"id":"c759b15f-131e-41d6-af3c-5680c8f1ea11"}

$ curl -X POST \
    -H 'Content-Type: application/json' \
    -d '{"id": "2", "name": "bookB", "detail": {"price": 200,"description": "this is sample book"}}' \
    http://localhost:8080/docs
{"id":"23a96578-e900-424f-a73f-808ff15d0823"}

$ curl -s http://localhost:8080/docs/23a96578-e900-424f-a73f-808ff15d0823 | jq
{
  "detail": {
    "description": "this is sample book",
    "price": 200
  },
  "id": "2",
  "name": "bookB"
}

$ curl --get -s http://localhost:8080/docs --data-urlencode 'q=name:"bookA"' | jq
{
  "count": 1,
  "documents": [
    {
      "document": {
        "detail": {
          "description": "this is sample book",
          "price": 100
        },
        "id": "1",
        "name": "bookA"
      },
      "id": "c759b15f-131e-41d6-af3c-5680c8f1ea11"
    }
  ]
}

$ curl --get -s http://localhost:8080/docs --data-urlencode 'q=detail.price:>150' | jq
{
  "count": 1,
  "documents": [
    {
      "document": {
        "detail": {
          "description": "this is sample book",
          "price": 200
        },
        "id": "2",
        "name": "bookB"
      },
      "id": "23a96578-e900-424f-a73f-808ff15d0823"
    }
  ]
}

$ curl --get -s http://localhost:8080/docs --data-urlencode 'q=detail.description:"this is sample book"' | jq
{
  "count": 2,
  "documents": [
    {
      "document": {
        "detail": {
          "description": "this is sample book",
          "price": 100
        },
        "id": "1",
        "name": "bookA"
      },
      "id": "c759b15f-131e-41d6-af3c-5680c8f1ea11"
    },
    {
      "document": {
        "detail": {
          "description": "this is sample book",
          "price": 200
        },
        "id": "2",
        "name": "bookB"
      },
      "id": "23a96578-e900-424f-a73f-808ff15d0823"
    }
  ]
}
```
