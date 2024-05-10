import requests
import time

def send_request():
    url = "http://your_vps_ip_address:10000"  # Replace "your_vps_ip_address" with your VPS's IP address
    try:
        response = requests.get(url)
        if response.status_code == 200:
            print("Response from server:", response.text)
            return response.text
        else:
            print("Error:", response.status_code)
    except requests.exceptions.RequestException as e:
        print("Error:", e)

def main():
    initial_response = send_request()
    if initial_response == "Hello, World!":
        print("Received 'Hello, World!' from server. Waiting for 'Goodbye, World!'...")
        time.sleep(10)  # Wait for 10 seconds
        goodbye_response = send_request()
        if goodbye_response == "Goodbye, World!":
            print("Received 'Goodbye, World!' from server.")
        else:
            print("Unexpected response from server after 10 seconds:", goodbye_response)
    else:
        print("Unexpected response from server:", initial_response)

if __name__ == "__main__":
    main()
