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
4. **Clients**: To use clients you need to clone the [Minislack-Client repository](https://github.com/wfercanas/Minislack-Client) and run the client following the instructions in the [README](https://github.com/wfercanas/Minislack-Client/blob/main/README.md).
5. Once your server is up and running, you will get to see the log whenever a user executes a command.

## Custom Protocol

For this project to work, I have created a custom protocol to control all the activity between the server and the clients. The protocol is designed on top of TCP and is text-based.

### Protocol commands

1. REG
2. JOIN
3. LEAVE
4. MSG
5. FILES
6. SEND
7. GET
8. CHNS
9. USRS

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
- When a message is sent to a channel, the channel then broadcasts it to all its joined users.

### FILES

**The structure of this command is: `FILES #channel`**

- This command is used to ask the server about files stored in a specific channel.
- The server answers with the list of files stored within the channel.

### SEND

**The structure of this command is: `SEND #channel <filename> <path>`**

- This command is used to send and save a file in a specific channel.
- The `<filename>` part of the command is the name you want to use to save a file. Is like using the "Save as.." button in the web.
- If the name you are trying to use isn't available, you will get an error message asking to try a different one.
- The `<path>` part of the command is the relative or absolute path to the file you want to save in the channel. Try using `./tmp/data.txt` or `./tmp/check.png` for testing.

### GET

**The structure of this command is: `GET #channel <filename>`**

- This command is used to ask for a file that is stored in a specific channel.
- The file is stored by default in the `./downloads/` folder of the client.
- If the client already has a file with that name in the `./downloads/` folder, the Client will handle it by appending an incremental number to differentiate them.

### CHNS

**The structure of this command is: `CHNS`**

This command is used to receive (the client) all the existing channels in the server.

### USRS

**The structure of this command is: `USRS`**

This command is used to receive (the client) all the users currently registered in the server

## Acknowledgements

Thanks to Ilija Eftimov for his article [Understanding Bytes in Go by building a TCP protocol](https://ieftimov.com/post/understanding-bytes-golang-build-tcp-protocol/). It was pretty helpful for creating this minislack project, mainly an extension of his work.
