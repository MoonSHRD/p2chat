## Package: api

### Imports:

None

### External Data, Input Sources:

None

### BaseMessage:

The `BaseMessage` struct represents the basic message format of the protocol. It has the following fields:

- `Body`: The message body.
- `To`: The recipient of the message.
- `Flag`: An integer representing the message type.
- `FromMatrixID`: The sender's MatrixID.

### GetTopicsRespondMessage:

The `GetTopicsRespondMessage` struct is used to respond to a request for existing PubSub topics at the network. It inherits from the `BaseMessage` struct and has an additional field:

- `Topics`: A list of strings representing the available PubSub topics.

The `Flag` field for this message type is set to 0x2.

