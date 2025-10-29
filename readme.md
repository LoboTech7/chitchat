# ChitChat Run Instructions

## 1. How to Test the System

### Step 1: Start the Server First
Open a terminal window, change the directory to the `server` folder, and run `go run .`. The server will start listening on port **8080**.

### Step 2: Start the Clients
Open at two or more new terminal windows, change the directory to the `client` folder, and run the client using `go run .`. Each client will connect to the server.

Once each client connected:
- You can type any message and send it. All connected clients should display it.
- You can type `.quit` in one client to leave the chat; the others should display that the participant has left the chat.

### Step 3. Leave Chat and Close Server
- Type `.quit` in all clients.
- To stop the server, use **Ctrl + C** in the server terminal.

## 2. Output and Logs
- Each client terminal displays received messages and timestamps.  
- The server records all major events in the **log.txt** file, including start server, join, leave, and message broadcasts.
