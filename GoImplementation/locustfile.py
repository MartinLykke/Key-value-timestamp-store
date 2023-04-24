from locust import HttpUser, task, between

class KeyValueTimestampUser(HttpUser):
    wait_time = between(0.5, 3.0)

    @task(1)
    def put_key_value_timestamp(self):
        self.client.put("/", json={"key": "test_key", "value": "test_value", "timestamp": 123456789})

    @task(2)
    def get_key_value_timestamp(self):
        self.client.get("/?key=test_key&timestamp=123456789")
