package operator

import (
	"context"
	"fmt"
	"log"

	pb "github.com/galxe/spotted-network/proto"
	"google.golang.org/grpc"
)

type RegistryClient struct {
    client pb.RegistryClient
    conn   *grpc.ClientConn
}

func NewRegistryClient(address string) (*RegistryClient, error) {
    log.Printf("Connecting to registry at %s", address)
    conn, err := grpc.Dial(address, grpc.WithInsecure())
    if err != nil {
        return nil, fmt.Errorf("failed to connect: %v", err)
    }
    log.Printf("Successfully connected to registry")
    
    return &RegistryClient{
        client: pb.NewRegistryClient(conn),
        conn:   conn,
    }, nil
}

func (c *RegistryClient) Close() {
    if c.conn != nil {
        log.Printf("Closing connection to registry")
        c.conn.Close()
    }
}

func (c *RegistryClient) GetRegistryID(ctx context.Context) (string, error) {
    log.Printf("Getting registry ID...")
    resp, err := c.client.GetRegistryID(ctx, &pb.GetRegistryIDRequest{})
    if err != nil {
        return "", fmt.Errorf("failed to get registry ID: %v", err)
    }
    log.Printf("Got registry ID: %s", resp.RegistryId)
    return resp.RegistryId, nil
}

func (c *RegistryClient) Join(ctx context.Context, address, message, signature string) (bool, error) {
    log.Printf("Submitting join request for address %s", address)
    resp, err := c.client.Join(ctx, &pb.JoinRequest{
        Address:   address,
        Message:   message,
        Signature: signature,
    })
    if err != nil {
        return false, fmt.Errorf("failed to submit join request: %v", err)
    }
    if !resp.Success {
        log.Printf("Join request failed: %s", resp.Error)
        return false, fmt.Errorf(resp.Error)
    }
    log.Printf("Join request successful")
    return true, nil
} 