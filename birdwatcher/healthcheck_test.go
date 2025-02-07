package birdwatcher

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck_addPrefix(t *testing.T) {
	hc := HealthCheck{}
	assert.Nil(t, hc.prefixes)

	// adding a prefix should initialise the prefixcollection
	// and add the prefix under the right prefixset
	_, prefix, _ := net.ParseCIDR("1.2.3.0/24")
	hc.addPrefix("foo", *prefix)
	assert.Equal(t, 1, len(hc.prefixes))
	assert.Equal(t, *prefix, hc.prefixes["foo"].prefixes[0])

	_, prefix, _ = net.ParseCIDR("2.3.4.0/24")
	hc.addPrefix("bar", *prefix)
	assert.Equal(t, 2, len(hc.prefixes))
	assert.Equal(t, *prefix, hc.prefixes["bar"].prefixes[0])
}

func TestHealthCheck_removePrefix(t *testing.T) {
	hc := HealthCheck{}
	assert.Nil(t, hc.prefixes)
	_, prefix, _ := net.ParseCIDR("1.2.3.0/24")
	hc.addPrefix("foo", *prefix)
	assert.Equal(t, 1, len(hc.prefixes))
	assert.Equal(t, 1, len(hc.prefixes["foo"].prefixes))

	// this should initialise the prefixset but won't remove any prefixes
	hc.removePrefix("bar", *prefix)
	assert.Equal(t, 2, len(hc.prefixes))
	assert.Equal(t, 1, len(hc.prefixes["foo"].prefixes))
	assert.Equal(t, 0, len(hc.prefixes["bar"].prefixes))

	// remove the prefix from the right prefixset
	hc.removePrefix("foo", *prefix)
	assert.Equal(t, 0, len(hc.prefixes["foo"].prefixes))
}

func TestHealthCheckDidReloadBefore(t *testing.T) {
	hc := NewHealthCheck(Config{})

	// expect both to fail
	assert.False(t, hc.didReloadBefore())

	// should succeed now
	hc.reloadedBefore = true
	assert.True(t, hc.didReloadBefore())

	hc.reloadedBefore = false

	// expect to fail again
	assert.False(t, hc.didReloadBefore())
}

func TestHealthCheck_handleAction(t *testing.T) {
	// empty healthcheck
	hc := HealthCheck{}
	assert.Nil(t, hc.prefixes)

	// create action with state up and 2 prefixes
	action := &Action{
		State:    ServiceStateUp,
		Prefixes: make([]net.IPNet, 2),
	}
	var prefix *net.IPNet
	_, prefix, _ = net.ParseCIDR("1.2.3.0/24")
	action.Prefixes[0] = *prefix
	_, prefix, _ = net.ParseCIDR("2.3.4.0/24")
	action.Prefixes[1] = *prefix
	action.Service = &ServiceCheck{
		FunctionName: "test",
	}
	// handle service state up
	hc.handleAction(action, nil)
	if assert.Contains(t, hc.prefixes, "test") {
		assert.Equal(t, 2, len(hc.prefixes["test"].prefixes))
	}

	// action switches to down for one of the prefixes
	action.State = ServiceStateDown
	action.Prefixes = action.Prefixes[1:]
	hc.handleAction(action, nil)
	if assert.Contains(t, hc.prefixes, "test") {
		assert.Equal(t, 1, len(hc.prefixes["test"].prefixes))
	}
}

func TestHealthCheck_statusUpdate(t *testing.T) {
	// healthcheck with 2 empty services
	hc := HealthCheck{services: []*ServiceCheck{
		{name: "foo"}, {name: "bar"},
	}}

	assert.Equal(t, "all 2 service(s) down", hc.statusUpdate())
	hc.services[0].state = ServiceStateUp
	assert.Equal(t, "service(s) bar down, 1 service(s) up", hc.statusUpdate())
	hc.services[1].state = ServiceStateUp
	assert.Equal(t, "all 2 service(s) up", hc.statusUpdate())
}
