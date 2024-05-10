import asyncio
import websockets

async def receive_messages():
    uri = "ws://captain-hook-server.onrender.com/ws"  # Replace "your.domain.com/ws" with the actual WebSocket URL

    async with websockets.connect(uri) as websocket:
        while True:
            message = await websocket.recv()
            print("Received message:", message)

asyncio.run(receive_messages())