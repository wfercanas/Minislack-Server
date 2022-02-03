# Mini Slack

## Project Description

This project creates a Slack version for the Command Line where users can register themselves, join #channels and send Direct Messages(DM).

## Technical Details

The project is built by:

1. Creating a basic TCP Custom Protocol for structuring the commands given by the clients.
2. Parsing and handling each command in a concurrent server working with goroutines.
3. Client and Server structs based on go channels for managing the flow of data.
4. Logging all activity in the Server and being verbose with the Client when using commands.

## How to use

1. First of all you need to have Go installed in your machine. For info about this point go to the [Download](https://go.dev/dl/) page of the official site.
2. Clone this repo in your computer: `git clone`
3. **Server**: Open a session in your terminal, go to the project folder and run the server: `go run .`. This will create a server running in local _(127.0.0.1)_ and listening the port _3000_.
4. **Clients**: For each client you must open a new session in your terminal. There, use `telnet 127.0.0.1 3000` to create a TCP connection with the server. You should receive a welcome message from the server.
5. Now you're free to start using MiniSlack. For more information about how to command instructions, see the Custom Protocol section below.

## Custom Protocol

For this project to work, I have created a custom protocol to control all the activity between the server and the clients. The protocol is designed on top of TCP and is text-based.

### Protocol commands

1. REG
2. JOIN
3. LEAVE
4. MSG
5. CHNS
6. USRS

### REG

**The structure of this command is: `REG @username`**.

- This command is used to let a user register himself in the server.
- Not an action will be allowed by the server if the user isn't registered.
- The server returns a successful response if that username value is available.

### JOIN

**The structure of this command is: `JOIN #channel`**.

- This command is used to let a user join a channel for writing and receving messages through it.
- If the channel the user is trying to join doesn't exist, the server will create it and automatically will join him/her.

### LEAVE

**The structure of this command is: `LEAVE #channel`**

- This command is used to let a user leave a channel, stopping more incoming messages to flow or the user to send new ones.
- If the channel doesn't exist, the server will send the client a message with a failure description.

### MSG

**The structure of this command is: `MSG #channel <length>\r\n<body>` or `MSG @username <length>\r\n<body>`**

- This command is used to send messages to channel or direct messages to users.
- The `<length>` part of the command must be equal to the length of the message body.
- The carriage return `\r\n` is a delimiter used before the body.
- The `<body>` is the message itself.
- When a message is sent to a channel, the channel then broadcast it to all its joined users.

### CHNS

**The structure of this command is: `CHNS`**

This command is used to receive (the client) all the existing channels in the server.

### USRS

**The structure of this command is: `USRS`**

This command is used to receive (the client) all the users currently registered in the server

## Acknowledgements

Thanks to Ilija Eftimov for his article [Understanding Bytes in Go by building a TCP protocol](https://ieftimov.com/post/understanding-bytes-golang-build-tcp-protocol/). A huge part of this project is based in his work there.
