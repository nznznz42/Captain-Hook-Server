import requests
import time

def send_request_and_print_response(url):
    response = requests.get(url)
    print("Response from", url, ":", response.text)

def main():
    url = "https://captain-hook-server.onrender.com"  # Replace this with the domain you want to send requests to
    send_request_and_print_response(url)
    print("Waiting for 60 seconds...")
    time.sleep(60)

if __name__ == "__main__":
    main()
