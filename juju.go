package juju

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/juju/juju/api"
	"github.com/juju/juju/api/client/client"
	"github.com/juju/juju/api/client/modelmanager"
	"github.com/juju/juju/api/connector"
	"github.com/juju/names/v5"
)

var log = clog.NewWithPlugin("juju")
var zones = []string{"juju.local."}

const ()

type Juju struct {
	Next        plugin.Handler
	Controllers map[string]Controller
	Ttl         uint32
}

func NewJuju(controllers map[string]Controller) *Juju {
	return &Juju{
		Controllers: controllers,
	}
}

func (j *Juju) GetAddress(ctx context.Context, fqdn JujuFQDN) (net.IP, error) {
	if !fqdn.IsValid() {
		return nil, fmt.Errorf("invalid juju FQDN %v", fqdn)
	}
	controller, ok := j.Controllers[fqdn.Controller]
	if !ok {
		return nil, fmt.Errorf("unknown controller %q", fqdn.Controller)
	}
	controllerConn, err := j.GetConnection(fqdn.Controller, "", controller)
	if err != nil {
		return nil, fmt.Errorf("getting connection for controller %q: %w", fqdn.Controller, err)
	}
	modelTag, err := j.GetModelTag(controllerConn, controller.Username, fqdn)

	return j.GetUnitAddress(ctx, controller, modelTag, fqdn)
}

func (j *Juju) GetUnitAddress(ctx context.Context, controller Controller, modelTag names.ModelTag, fqdn JujuFQDN) (net.IP, error) {
	modelConn, err := j.GetConnection(fqdn.Controller, modelTag.Id(), controller)
	if err != nil {
		return nil, fmt.Errorf("getting connection for model %q: %w", modelTag.Id(), err)
	}
	apiClient := client.NewClient(modelConn, nil)
	fullStatus, err := apiClient.Status(&client.StatusArgs{})
	if err != nil {
		return nil, fmt.Errorf("getting status for model %q: %w", fqdn.Model, err)
	}
	app, ok := fullStatus.Applications[fqdn.Application]
	if !ok {
		return nil, fmt.Errorf("unknown application %q", fqdn.Application)
	}
	unit, ok := app.Units[fqdn.Application+"/"+fqdn.Unit]
	if !ok {
		return nil, fmt.Errorf("unknown unit %q", fqdn.Unit)
	}
	return net.ParseIP(unit.PublicAddress), nil
}

func (j *Juju) GetConnection(controllerName, modelUUID string, controller Controller) (api.Connection, error) {
	opts := connector.SimpleConfig{
		ControllerAddresses: []string{controller.address},
		Username:            controller.Username,
		Password:            controller.Password,
	}
	if modelUUID != "" {
		opts.ModelUUID = modelUUID
	}
	simpleConnector, err := connector.NewSimple(opts, api.WithDialOpts(api.DialOpts{
		Timeout:            time.Millisecond * 400,
		InsecureSkipVerify: true,
	}))
	if err != nil {
		return nil, fmt.Errorf("creating simple connector for controller %q and model %q: %w", controllerName, modelUUID, err)
	}
	conn, err := simpleConnector.Connect()
	if err != nil {
		return nil, fmt.Errorf("connecting to controller %q and model %q: %w", controllerName, modelUUID, err)
	}
	return conn, nil
}

func (j *Juju) GetModelTag(conn api.Connection, username string, fqdn JujuFQDN) (names.ModelTag, error) {
	modelMgr := modelmanager.NewClient(conn)
	models, err := modelMgr.ListModelSummaries(username, true)
	if err != nil {
		return names.ModelTag{}, fmt.Errorf("listing models for controller %q: %w", fqdn.Controller, err)
	}
	var modelTag names.ModelTag
	for _, model := range models {
		if model.Name == fqdn.Model {
			modelTag = names.NewModelTag(model.UUID)
			break
		}
	}
	if modelTag.Id() == "" {
		return names.ModelTag{}, fmt.Errorf("unknown model %q", fqdn.Model)
	}
	return modelTag, nil
}

// Name implements the Handler interface.
func (j *Juju) Name() string { return "juju" }
