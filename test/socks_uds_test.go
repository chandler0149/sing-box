package main

import (
	"net"
	"path/filepath"
	"testing"
	"time"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/stretchr/testify/require"
)

func TestSocksUDS(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "socks.sock")

	// Start sing-box instance
	startInstance(t, option.Options{
		Inbounds: []option.Inbound{
			{
				Type: C.TypeSOCKS,
				Tag:  "socks-uds-in",
				Options: &option.SocksInboundOptions{
					ListenOptions: option.ListenOptions{
						ListenPath: socketPath,
					},
				},
			},
		},
		Outbounds: []option.Outbound{
			{
				Type: C.TypeBlock,
				Tag:  "block-out",
			},
		},
	})

	// Wait briefly for the server to start and bind
	time.Sleep(200 * time.Millisecond)

	// Dial the UDS socket directly
	conn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	defer conn.Close()

	// Write SOCKS5 greeting: Version 5, 1 auth method: No Auth (0)
	_, err = conn.Write([]byte{5, 1, 0})
	require.NoError(t, err)

	// Read greeting response
	buf := make([]byte, 2)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, byte(5), buf[0])
	require.Equal(t, byte(0), buf[1])

	// Send CONNECT command:
	// Version: 5, CMD: Connect (1), RSV: 0
	// ATYP: Domain (3), Length: 10, Domain: "google.com", Port: 80 (0x00, 0x50)
	commandBytes := []byte{5, 1, 0, 3, 10, 'g', 'o', 'o', 'g', 'l', 'e', '.', 'c', 'o', 'm', 0, 80}

	_, err = conn.Write(commandBytes)
	require.NoError(t, err)

	// Read connect response
	resBuf := make([]byte, 1024)
	n, err := conn.Read(resBuf)
	require.NoError(t, err)
	require.True(t, n >= 10)
	require.Equal(t, byte(5), resBuf[0])
	// SOCKS5 reply field (resBuf[1]) represents reply status: either success (0) or error/blocked (non-zero).
	t.Logf("SOCKS5 connect response code: %d", resBuf[1])
}
