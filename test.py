import requests
from http.server import HTTPServer, BaseHTTPRequestHandler
import threading
from urllib.parse import urlparse

# Define a simple HTTP request handler
class RequestHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        print("Received request from server:")
        print(post_data.decode('utf-8'))

# Function to start the HTTP server in a separate thread
def start_server(host, port):
    server_address = (host, port)
    httpd = HTTPServer(server_address, RequestHandler)
    print(f"Server started on {host}:{port}")
    httpd.serve_forever()

# Function to send a request to a domain and print the response
def send_request(url):
    parsed_url = urlparse(url)
    domain_host = parsed_url.hostname
    domain_port = parsed_url.port or 80
    
    response = requests.get(url)
    print("Response from server:")
    print(response.text)
    
    # Start the HTTP server in a separate thread
    server_thread = threading.Thread(target=start_server, args=(domain_host, domain_port))
    server_thread.daemon = True
    server_thread.start()
    
    print("Server initiated, waiting for requests...")

# Main function
def main():
    domain_url = "https://captain-hook-server.onrender.com"
    print("Sending request to", domain_url)
    send_request(domain_url)

if __name__ == "__main__":
    main()
