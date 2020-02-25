import threading
from simple_websocket_server import WebSocketServer

from config import WEB_SOCKET_CONFIG
from data.factory import Factory
from web_socket.web_socket_handler import WebSocketHandler
from web_socket.web_socket_message_handler import WebSocketMessageHandler


class Hp2pWebSocketServer:
    def __init__(self):
        self.host = WEB_SOCKET_CONFIG['HOST']
        self.port = WEB_SOCKET_CONFIG['PORT']

    def start(self):
        t = threading.Thread(target=self.run_web_socket_server, daemon=True)
        t.start()

    def run_web_socket_server(self):
        print("[SERVER] Start Web Socket Server...")

        handler = WebSocketMessageHandler()
        Factory.get().set_web_socket_message_handler(handler)

        server = WebSocketServer(self.host, self.port, WebSocketHandler)
        server.serve_forever()
