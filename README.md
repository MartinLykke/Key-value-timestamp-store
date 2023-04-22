# Key-value-timestamp-store

Install dependencies:
```
pip install flask
```

run the api with:
```
python key_value_timestamp_store.py
```

# How to PUT and GET
```
curl -X PUT http://127.0.0.1:5000 -H "Content-Type: application/json" -d "{\"key\": \"mykey\", \"value\": \"myvalue\", \"timestamp\": 1673524092123456}"

curl -X GET "http://127.0.0.1:5000?key=mykey&timestamp=1673524092123456
```
# How to run a load test

```
pip install locust

locust
```

Then go to http://localhost:8089/
