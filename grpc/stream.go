package grpc

// type Client struct {
// 	fp Freedom_PipeClient
// }

// func (c *Client) Read(b []byte) (int, error) {

// 	req, err := c.fp.Recv()
// 	if err != nil {
// 		return 0, err
// 	}
// 	data := req.GetData()
// 	b = data
// 	return len(data), nil
// }

// func (c *Client) Write(b []byte) (int, error) {
// 	length := len(b)
// 	err := c.fp.Send(&FreedomRequest{Data: b})
// 	return length, err
// }

// func (c *Client) Close() error {
// 	return c.fp.CloseSend()
// }
