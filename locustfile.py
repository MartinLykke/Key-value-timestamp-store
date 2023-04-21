from locust import HttpUser, task, between

class MyUser(HttpUser):
    wait_time = between(1, 5)

    @task
    def put_data(self):
        headers = {'Content-Type': 'application/json'}
        data = {"key": "mykey", "value": "myvalue", "timestamp": 1673524092123456}
        self.client.put("/", json=data, headers=headers)

    @task
    def get_data(self):
        params = {'key': 'mykey', 'timestamp': 1673524092123456}
        self.client.get("/", params=params)
