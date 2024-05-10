from http.server import BaseHTTPRequestHandler, HTTPServer
import threading

class RequestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/hello":
            self.send_response(200)
            self.send_header("Content-type", "text/plain")
            self.end_headers()
            self.wfile.write(b"Hello, world!")

class GoServerConnection:
    def __init__(self):
        self.go_program_connection = None

def run_server(go_server_connection):
    server_address = ('', 8000)
    httpd = HTTPServer(server_address, RequestHandler)
    go_server_connection.go_program_connection = httpd
    print("Starting server...")
    httpd.serve_forever()

def main():
    go_server_connection = GoServerConnection()

    # Start the Python HTTP server
    server_thread = threading.Thread(target=run_server, args=(go_server_connection,))
    server_thread.start()

    # Wait for the Go program to initiate further communication
    input("Press Enter to exit...")

if __name__ == "__main__":
    main()
