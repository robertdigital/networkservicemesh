package vppagent

import (
	"context"

	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/crossconnect"
	"github.com/networkservicemesh/networkservicemesh/forwarder/api/forwarder"
	"github.com/networkservicemesh/networkservicemesh/forwarder/vppagent/pkg/converter"
)

type clearMechanisms struct {
	baseDir string
}

//ClearMechanisms sends clear datachange request if crossconnect monitor has entity with request cross conenect id.
func ClearMechanisms(baseDir string) forwarder.ForwarderServer {
	return &clearMechanisms{
		baseDir: baseDir,
	}
}

func (c *clearMechanisms) Request(ctx context.Context, crossConnect *crossconnect.CrossConnect) (*crossconnect.CrossConnect, error) {
	conversionParameters := &converter.CrossConnectConversionParameters{
		BaseDir: c.baseDir,
	}
	monitor := MonitorServer(ctx)
	if monitor == nil {
		logrus.Info("Crossconnect monitor server not passed")
		return nextRequest(ctx, crossConnect)
	}
	entity := monitor.Entities()[crossConnect.GetId()]
	if entity == nil {
		logrus.Infof("montir has not entry with id %v", crossConnect.GetId())
		return nextRequest(ctx, crossConnect)
	}
	clearDataChange, err := converter.NewCrossConnectConverter(entity.(*crossconnect.CrossConnect), conversionParameters).MechanismsToDataRequest(nil, false)
	if err == nil && clearDataChange != nil {
		logrus.Infof("Sending clearing DataChange to vppagent: %v", proto.MarshalTextString(clearDataChange))
		client := ConfiguratorClient(ctx)
		if client == nil {
			return nil, errors.New("configuration client is not passed for clear mechanism")
		}
		_, err = client.Delete(ctx, &configurator.DeleteRequest{Delete: clearDataChange})
	}
	if err != nil {
		logrus.Warnf("Connection Mechanism was not cleared properly before updating: %s", err.Error())
	}
	return nextRequest(ctx, crossConnect)
}

func (c *clearMechanisms) Close(ctx context.Context, crossConnect *crossconnect.CrossConnect) (*empty.Empty, error) {
	if next := Next(ctx); next != nil {
		return next.Close(ctx, crossConnect)
	}
	return new(empty.Empty), nil
}

func nextRequest(ctx context.Context, crossConnect *crossconnect.CrossConnect) (*crossconnect.CrossConnect, error) {
	if next := Next(ctx); next != nil {
		return next.Request(ctx, crossConnect)
	}
	return crossConnect, nil
}
