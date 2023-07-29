package test

// import (
// 	"context"
// 	"net"
// 	"strings"
// 	"time"
// )

// const (
// 	waitPortInterval    = 100 * time.Millisecond
// 	waitPortConnTimeout = 50 * time.Millisecond
// )

// func WaitPort(ctx context.Context, network, port string) error {
// 	ticker := time.NewTicker(waitPortInterval)
// 	defer ticker.Stop()

// 	port = strings.TrimLeft(port, ":")

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		case <-ticker.C:
// 			conn, _ := net.DialTimeout(network, ":"+port, waitPortConnTimeout)
// 			if conn != nil {
// 				_ = conn.Close()
// 				return nil
// 			}
// 		}
// 	}
// }
