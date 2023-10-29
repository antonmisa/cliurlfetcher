package integration_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	constFileName    = "data.csv"
	constLastMessage = "DONE"
)

type TestLogConsumer struct {
	Msgs []string
	Done chan bool
}

func NewTestLogConsumer() *TestLogConsumer {
	return &TestLogConsumer{
		Done: make(chan bool),
	}
}

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	s := string(l.Content)

	if strings.Contains(s, constLastMessage) {
		g.Done <- true
		return
	}

	g.Msgs = append(g.Msgs, s)
}

type nginxContainer struct {
	testcontainers.Container
	URI string
}

type cmdContainer struct {
	testcontainers.Container
	file string
}

type localNetwork struct {
	testcontainers.Network
	name string
}

func setupNetwork(ctx context.Context) (*localNetwork, error) {
	networkName := "cliutlfetcher-network"

	net, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	})

	if err != nil {
		return nil, err
	}

	return &localNetwork{Network: net, name: networkName}, nil
}

func setupNginx(ctx context.Context, net *localNetwork) (*nginxContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
		Privileged:   true,
		Networks: []string{
			net.name,
		},
		NetworkAliases: map[string][]string{
			net.name: []string{"cmd", "nginx"},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "80")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	return &nginxContainer{Container: container, URI: uri}, nil
}

func genData() [][2]string {
	res := make([][2]string, 0, 2)

	res = append(res, [2]string{"http://nginx", "200"})
	res = append(res, [2]string{"http://nginx/notexist", "404"})

	return res
}

func createFile(data [][2]string) error {
	f, err := os.Create(constFileName)
	if err != nil {
		return err
	}

	defer f.Close()

	for _, e := range data {
		_, err = f.WriteString(fmt.Sprintf("%s\n", e[0]))
		if err != nil {
			return err
		}
	}

	return nil
}

func analyzeFile(data [][2]string, msgs string) error {
	for _, e := range data {
		s := fmt.Sprintf("Completed url: %s, status: %s,", e[0], e[1])

		if !strings.Contains(msgs, s) {
			return fmt.Errorf("incorrect final result for: %s", s)
		}
	}

	return nil
}

func deleteFile() error {
	err := os.Remove(constFileName)
	if err != nil {
		return err
	}

	return nil
}

func setupLocal(ctx context.Context, net *localNetwork) (*cmdContainer, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    "../.",
				Dockerfile: "./integration-test/Dockerfile",
			},
			Privileged: true,
			Networks: []string{
				net.name,
			},
			NetworkAliases: map[string][]string{
				net.name: []string{"cmd", "nginx"},
			},
			WaitingFor: wait.ForExit(),
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      filepath.Join("./", constFileName),
					ContainerFilePath: filepath.Join("/", constFileName),
					FileMode:          0o444,
				},
			},
		},
		Started: false,
	})
	if err != nil {
		return nil, err
	}

	return &cmdContainer{Container: container, file: constFileName}, nil
}

func TestIntegrationNginxReturn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	data := genData()

	net, err := setupNetwork(ctx)
	require.NoError(t, err)

	nginxC, err := setupNginx(ctx, net)
	require.NoError(t, err)

	err = createFile(data)
	require.NoError(t, err)

	cmdC, err := setupLocal(ctx, net)
	require.NoError(t, err)

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		_ = deleteFile()

		err := cmdC.Terminate(ctx)
		require.NoError(t, err)

		err = nginxC.Terminate(ctx)
		require.NoError(t, err)

		err = net.Remove(ctx)
		require.NoError(t, err)
	})

	//tlc := TestLogConsumer{
	//	Msgs: []string{},
	//}

	//cmdC.FollowOutput(&tlc) // must be called before StarLogProducer

	//err = cmdC.StartLogProducer(ctx)
	//require.NoError(t, err)

	err = cmdC.Start(ctx)
	require.NoError(t, err)

	time.Sleep(time.Second * 5)

	r, err := cmdC.Logs(ctx)
	require.NoError(t, err)

	defer r.Close()

	b, err := io.ReadAll(r)
	require.NoError(t, err)

	//	require.Equal(t, string(b), "")

	err = analyzeFile(data, string(b))
	require.NoError(t, err)
}
