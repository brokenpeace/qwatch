package qinput

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
    "github.com/zpatrick/go-config"

	"github.com/qnib/qwatch/types"
)
// DockerEvents is a simple qworker
type DockerEvents struct {
    qtypes.QWorker
}

// NewDockerEvents returns instance of DockerEventInput
func NewDockerEvents(cfg *config.Config, qC qtypes.Channels) DockerEvents {
    de := DockerEvents{}
    de.Cfg = cfg
    de.QChan = qC
    return de
}
// Run subscribes to messages and events from the docker-engine
func (de DockerEvents) Run() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	msgs, errs := cli.Events(context.Background(), types.EventsOptions{})
	bg := de.QChan.Log.Join()
	for {
		select {
		case dMsg := <-msgs:
			bg.Send(parseMessage(dMsg))
		case dErr := <-errs:
			if dErr != nil {
				qm := qtypes.Qmsg{
					Msg: fmt.Sprintf("%s", dErr),
				}
				bg.Send(qm)
			}
		}
	}
}

func parseMessage(msg events.Message) qtypes.Qmsg {
	host := os.Getenv("DOCKER_HOST")
	message := fmt.Sprintf("%s.%s", msg.Type, msg.Action)
	cnt := qtypes.ContainerInfo{
		ImageName:     msg.Actor.Attributes["image"],
		ContainerID:   msg.ID,
		ContainerName: msg.Actor.Attributes["name"],
	}
	qm := qtypes.Qmsg{
		Version:     "1.1",
		Source:      "docker-events",
		Host:        host,
		Msg:         message,
		IsContainer: false,
		Time:        time.Unix(0, msg.TimeNano),
	}
	qm.SetContainer(cnt)
	qm.Type = fmt.Sprintf("%s.%s", msg.Type, msg.Action)
	return qm
}