from flask import Flask, request
import json
import os

app = Flask(__name__)

data = {}

@app.route('/', methods=['PUT'])
def put():
    try:
        # get input data from the request
        req_data = request.get_json()
        # extract the key, value, and timestamp from the input data
        key = req_data['key']
        value = req_data['value']
        timestamp = req_data['timestamp']
        # validate the data types of the input parameters
        if not isinstance(key, str) or not isinstance(value, str) or not isinstance(timestamp, int):
            raise ValueError("Invalid input data type")
        # add the key-value pair to the data dictionary
        if key not in data:
            data[key] = []
        data[key].append({'timestamp': timestamp, 'value': value})
        # write the key-value pair to a file
        with open('data.txt', 'a') as f:
            f.write(json.dumps({'key': key, 'value': value, 'timestamp': timestamp})+'\n')
        # return a success message
        return 'OK'
    except (KeyError, ValueError, json.JSONDecodeError):
        # return an error message if there was an error processing the request
        return 'Bad request', 400

@app.route('/', methods=['GET'])
def get():
    try:
        # extract the key and timestamp from the request parameters
        key = request.args.get('key')
        timestamp = int(request.args.get('timestamp'))
        # validate the data types of the input parameters
        if not isinstance(key, str) or not isinstance(timestamp, int):
            raise ValueError("Invalid input data type")
        # look up the key in the data dictionary
        if key not in data:
            return 'Key not found'
        # iterate over the list of key-value pairs for the given key
        for item in reversed(data[key]):
            # return the value for the latest key-value pair with a timestamp <= the given timestamp
            if item['timestamp'] <= timestamp:
                return item['value']
        # return a message indicating that no value was found
        return 'No value found'
    except ValueError:
        # return an error message if there was an error processing the request
        return 'Bad request', 400

if __name__ == '__main__':
    # load data from file if it exists
    if not os.path.exists('data.txt'):
        with open('data.txt', 'w') as f:
            f.write('')
    else:
        with open('data.txt', 'r') as f:
            # iterate over the lines in the file
            for line in f:
                try:
                    # parse each line as a JSON object
                    item = json.loads(line.strip())
                    # validate the data types of the key, value, and timestamp
                    if not isinstance(item['key'], str) or not isinstance(item['value'], str) or not isinstance(item['timestamp'], int):
                        raise ValueError("Invalid input data type")
                    # add the key-value pair to the data dictionary
                    if item['key'] not in data:
                        data[item['key']] = []
                    data[item['key']].append({'timestamp': item['timestamp'], 'value': item['value']})
                except (ValueError, json.JSONDecodeError):
                    # ignore lines that cannot be parsed as JSON
                    pass
    # start the Flask application
    app.run()
